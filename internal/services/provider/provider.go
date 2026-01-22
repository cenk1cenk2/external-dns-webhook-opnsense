package provider

import (
	"context"
	"fmt"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/opnsense"
	"go.uber.org/zap"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

type Provider struct {
	provider.BaseProvider

	Config ProviderConfig

	Log          services.ZapSugaredLogger
	Client       opnsense.ClientAdapter
	DomainFilter endpoint.DomainFilterInterface
}

type ProviderSvc struct {
	Client opnsense.ClientAdapter
	Logger *services.Logger
}

type ProviderConfig struct {
	DomainFilter DomainFilterConfig
}

var _ provider.Provider = (*Provider)(nil)

// NewProvider creates a new OPNsense DNS provider.
func NewProvider(svc *ProviderSvc, conf ProviderConfig) (*Provider, error) {
	return &Provider{
		Config:       conf,
		Client:       svc.Client,
		Log:          svc.Logger.WithCaller().With(zap.String("service", "provider")),
		DomainFilter: NewDomainFilter(conf.DomainFilter),
	}, nil
}

// Records returns the list of records from OPNsense Unbound DNS.
func (p *Provider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	result, err := p.Client.UnboundSearchHostOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query for host overrides: %w", err)
	}

	endpoints := make([]*endpoint.Endpoint, 0)

	for _, row := range result.Rows {
		record := NewDnsRecord(row)
		p.Log.Debugf("Processing record: %+v", record)

		if !p.GetDomainFilter().Match(record.GetFQDN()) {
			p.Log.Debugf("Skipping record due to domain filter: %s", record.GetFQDN())
			continue
		}

		ep := endpoint.
			NewEndpoint(
				record.GetFQDN(),
				record.Type,
				record.GetTarget()...,
			).
			WithLabel(EndpointLabelUUID.String(), record.Id)
		if ep == nil {
			return nil, fmt.Errorf("failed to create endpoint for record %s", record.GetFQDN())
		}

		ep.SetIdentifier = record.GenerateSetIdentifier()

		if record.Description != "" {
			ep.WithProviderSpecific(
				ProviderSpecificDescription.String(),
				record.Description,
			)
		}

		p.Log.Debugf("Endpoint processed: %+v", ep)

		endpoints = append(endpoints, ep)
	}

	return endpoints, nil
}

func (p *Provider) AdjustEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error) {
	// opnsense unbound dns doesn't support multiple targets per record.
	// split endpoints with multiple targets into separate endpoints with unique setidentifiers.
	adjusted := make([]*endpoint.Endpoint, 0, len(endpoints))

	for _, ep := range endpoints {
		p.Log.Debugf("AdjustEndpoints processing: %+v", ep)

		// Skip endpoints with no targets - nothing to create
		if len(ep.Targets) == 0 {
			p.Log.Debugf("Skipping endpoint with no targets: %s", ep.DNSName)
			continue
		}

		// If endpoint has only one target, no splitting needed - just ensure SetIdentifier is set
		if len(ep.Targets) == 1 {
			p.Log.Debugf("Endpoint has single target, no splitting needed: %s", ep.DNSName)

			// Create a record to generate SetIdentifier
			records, err := NewDnsRecordsFromEndpoint(ep)
			if err != nil {
				return nil, fmt.Errorf("failed to create record from endpoint %s: %v", ep.DNSName, err)
			}

			ep.SetIdentifier = records[0].GenerateSetIdentifier()
			ep.WithLabel(EndpointLabelSetIdentifier.String(), ep.SetIdentifier)

			adjusted = append(adjusted, ep)
			continue
		}

		// Multiple targets - need to split into separate endpoints
		records, err := NewDnsRecordsFromEndpoint(ep)
		if err != nil {
			return nil, fmt.Errorf("failed to create records from endpoint %s: %v", ep.DNSName, err)
		}

		p.Log.Debugf("Normalized endpoint %s into %d record(s)", ep.DNSName, len(records))

		for _, record := range records {
			if ep.Labels == nil {
				ep.Labels = make(map[string]string)
			}
			ep.Labels[EndpointLabelUUID.String()] = "" // Clear UUID for split endpoints

			e := &endpoint.Endpoint{
				DNSName:          ep.DNSName,
				Targets:          record.GetTarget(),
				RecordType:       ep.RecordType,
				SetIdentifier:    record.GenerateSetIdentifier(),
				RecordTTL:        ep.RecordTTL,
				Labels:           ep.Labels,
				ProviderSpecific: ep.ProviderSpecific,
			}

			e.WithLabel(EndpointLabelSetIdentifier.String(), e.SetIdentifier)
			// Note: UUID is intentionally not set here for multi-target splits
			// Each split endpoint will need to be matched against Records() to find its UUID

			adjusted = append(adjusted, e)
			p.Log.Debugf("Adjusted endpoint: %s -> %v", ep.DNSName, record.GetTarget())
		}
	}

	p.Log.Debugf("AdjustEndpoints returning %d endpoint(s)", len(adjusted))

	return adjusted, nil
}

// ApplyChanges applies a set of changes to OPNsense Unbound DNS.
func (p *Provider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	p.Log.Debugf("ApplyChanges called with %d creates, %d updates, %d deletes", len(changes.Create), len(changes.UpdateNew), len(changes.Delete))

	for _, ep := range changes.Delete {
		p.Log.Debugf("Delete request for: %+v", ep)

		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA:
			record, err := NewDnsRecordFromExistingEndpoint(ep)
			if err != nil {
				return fmt.Errorf("failed to create record from endpoint %s: %w", ep.DNSName, err)
			}

			if err := p.Client.UnboundDeleteHostOverride(ctx, record.Id); err != nil {
				return fmt.Errorf("failed to delete host override %s with correct UUID %s: %w", ep.DNSName, record.Id, err)
			}

			p.Log.Infof(
				"Deleted host override: %s (%s) with id %s, SetIdentifier: %s",
				ep.DNSName,
				ep.RecordType,
				record.Id,
				ep.SetIdentifier,
			)
		case endpoint.RecordTypeTXT:
			p.Log.Debugf("Processing TXT record delete: %s", ep.DNSName)

			record, err := p.handleTxtRecordMatching(ctx, ep)
			if err != nil {
				return fmt.Errorf("failed to match TXT record for delete: %w", err)
			}

			p.Log.Debugf("Found matching TXT record to delete: %s with UUID %s", record.GetFQDN(), record.Id)

			if err := p.Client.UnboundDeleteHostOverride(ctx, record.Id); err != nil {
				return fmt.Errorf("failed to delete TXT host override %s with UUID %s: %w", ep.DNSName, record.Id, err)
			}

			p.Log.Infof(
				"Deleted host override: %s (%s) with id %s, SetIdentifier: %s",
				ep.DNSName,
				ep.RecordType,
				record.Id,
				ep.SetIdentifier,
			)

		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", ep.RecordType, ep.DNSName)
		}
	}

	// UpdateOld and UpdateNew are parallel arrays with matching indices
	for i, newEp := range changes.UpdateNew {
		oldEp := changes.UpdateOld[i]
		p.Log.Debugf("Update request for: from %+v to %+v", oldEp, newEp)

		switch newEp.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA:
			oldRecord, err := NewDnsRecordFromExistingEndpoint(oldEp)
			if err != nil {
				return fmt.Errorf("failed to create record from existing endpoint %s: %w", oldEp.DNSName, err)
			}

			newRecord, err := NewDnsRecordFromEndpoint(newEp)
			if err != nil {
				return fmt.Errorf("failed to create record from endpoint %s: %w", newEp.DNSName, err)
			}
			newRecord.Id = oldRecord.Id

			p.Log.Debugf("Updating host override: %s (%s) with id %s", newEp.DNSName, newEp.RecordType, newRecord.Id)
			if err := p.Client.UnboundUpdateHostOverride(ctx, newRecord.Id, newRecord.IntoHostOverride()); err != nil {
				return fmt.Errorf("failed to update host override %s: %w", newEp.DNSName, err)
			}
			p.Log.Infof("Updated host override: %s (%s) with id %s", newEp.DNSName, newEp.RecordType, newRecord.Id)

		case endpoint.RecordTypeTXT:
			p.Log.Debugf("Processing TXT record update: %s", oldEp.DNSName)

			oldRecord, err := p.handleTxtRecordMatching(ctx, oldEp)
			if err != nil {
				return fmt.Errorf("failed to match TXT record for update: %w", err)
			}

			newRecord, err := NewDnsRecordFromEndpoint(newEp)
			if err != nil {
				return fmt.Errorf("failed to create record from endpoint %s: %w", newEp.DNSName, err)
			}

			newRecord.Id = oldRecord.Id

			p.Log.Debugf("Updating host override: %s (%s) with id %s", newEp.DNSName, newEp.RecordType, newRecord.Id)
			if err := p.Client.UnboundUpdateHostOverride(ctx, newRecord.Id, newRecord.IntoHostOverride()); err != nil {
				return fmt.Errorf("failed to update host override %s: %w", newEp.DNSName, err)
			}
			p.Log.Infof("Updated host override: %s (%s) with id %s", newEp.DNSName, newEp.RecordType, newRecord.Id)

		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", newEp.RecordType, newEp.DNSName)
		}
	}

	for _, ep := range changes.Create {
		p.Log.Debugf("Create request for: %+v", ep)

		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			records, err := NewDnsRecordsFromEndpoint(ep)
			if err != nil {
				return fmt.Errorf("failed to create records from endpoint %s: %w", ep.DNSName, err)
			}

			for _, record := range records {
				p.Log.Debugf("Creating host override: %s (%s) -> %+v", ep.DNSName, ep.RecordType, record.GetTarget())
				uuid, err := p.Client.UnboundCreateHostOverride(ctx, record.IntoHostOverride())
				if err != nil {
					return fmt.Errorf("failed to create host override %s: %w", ep.DNSName, err)
				}
				p.Log.Infof("Created host override: %s (%s) -> %+v, with id %s", ep.DNSName, ep.RecordType, record.GetTarget(), uuid)
			}

		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", ep.RecordType, ep.DNSName)
		}

	}

	return nil
}

// GetDomainFilter returns the domain filter for this provider.
func (p *Provider) GetDomainFilter() endpoint.DomainFilterInterface {
	return p.DomainFilter
}

// handleTxtRecordMatching handles the logic for finding the correct DnsRecord for a given registry TXT record.
func (p *Provider) handleTxtRecordMatching(ctx context.Context, ep *endpoint.Endpoint) (*DnsRecord, error) {
	record, err := NewDnsRecordFromEndpoint(ep)
	if err != nil {
		return nil, fmt.Errorf("failed to create record from endpoint %s: %w", ep.DNSName, err)
	}

	if _, err := endpoint.NewLabelsFromString(record.TxtData, nil); err == nil {
		p.Log.Debugf("Endpoint corresponds to a registry record: %+v", ep)
		all, err := p.Records(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all records for matching TXT record: %w", err)
		}

		var matched *endpoint.Endpoint
		for _, current := range all {
			if current.DNSName == ep.DNSName && current.RecordType == ep.RecordType && current.SetIdentifier == ep.SetIdentifier && current.Labels != nil && ep.Labels != nil &&
				current.Labels["owner"] == ep.Labels["owner"] && current.Labels["resource"] == ep.Labels["resource"] {
				matched = current
				break
			}
		}

		if matched == nil {
			return nil, fmt.Errorf("failed to find matching TXT record for %s with SetIdentifier %s", ep.DNSName, ep.SetIdentifier)
		}

		record, err := NewDnsRecordFromExistingEndpoint(matched)
		if err != nil {
			return nil, fmt.Errorf("failed to create record from matched endpoint %s: %w", ep.DNSName, err)
		}

		return record, nil
	}

	// Labels not parsable - this is a normal TXT record, use UUID from Labels directly
	p.Log.Debugf("Endpoint corresponds to a normal TXT record: %+v", ep)

	record, err = NewDnsRecordFromExistingEndpoint(ep)
	if err != nil {
		return nil, fmt.Errorf("failed to create record from endpoint %s: %w", ep.DNSName, err)
	}

	return record, nil
}
