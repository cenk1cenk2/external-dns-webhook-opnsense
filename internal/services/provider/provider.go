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
			WithProviderSpecific(
				ProviderSpecificUUID.String(),
				record.Id,
			)
		if ep == nil {
			return nil, fmt.Errorf("failed to create endpoint for record %s", record.GetFQDN())
		}

		// For TXT records, try to parse labels from TXT data to extract SetIdentifier
		if record.Type == endpoint.RecordTypeTXT {
			p.Log.Debugf("Processing TXT record: %s, TxtData: %s", record.GetFQDN(), record.TxtData)
			if labels, err := endpoint.NewLabelsFromString(record.TxtData, nil); err == nil {
				p.Log.Debugf("Successfully parsed labels from TXT data: %+v", labels)
				// This is a TXT registry record, extract labels into endpoint
				for k, v := range labels {
					ep.Labels["external-dns/"+k] = v
				}
				// Extract SetIdentifier from labels if it exists
				if setID, exists := labels["set-identifier"]; exists {
					p.Log.Debugf("Found set-identifier in TXT labels: %s", setID)
					ep.SetIdentifier = setID
				} else {
					// Not a registry record with SetIdentifier, generate one
					ep.SetIdentifier = record.GenerateSetIdentifier()
					p.Log.Debugf("No set-identifier in TXT labels, generated: %s", ep.SetIdentifier)
				}
			} else {
				// Not a registry record, generate SetIdentifier normally
				p.Log.Debugf("Failed to parse TXT data as labels (error: %v), generating SetIdentifier", err)
				ep.SetIdentifier = record.GenerateSetIdentifier()
				p.Log.Debugf("Generated SetIdentifier for non-registry TXT: %s", ep.SetIdentifier)
			}
		} else {
			// For non-TXT records, try to get SetIdentifier from labels
			if setIdentifier, exists := ep.Labels["external-dns/set-identifier"]; exists {
				p.Log.Debugf("Found set-identifier in %s record labels: %s", record.Type, setIdentifier)
				ep.SetIdentifier = setIdentifier
			} else {
				ep.SetIdentifier = record.GenerateSetIdentifier()
				p.Log.Debugf("No set-identifier in %s record labels, generated: %s", record.Type, ep.SetIdentifier)
			}
		}

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
		p.Log.Debugf("AdjustEndpoints processing: %s (%s), targets: %v, existing SetIdentifier: %s", ep.DNSName, ep.RecordType, ep.Targets, ep.SetIdentifier)

		// Keep endpoints that already have SetIdentifiers
		if ep.SetIdentifier != "" {
			adjusted = append(adjusted, ep)
			p.Log.Debugf("Keeping endpoint with existing SetIdentifier: %s", ep.SetIdentifier)
			continue
		}

		// Skip endpoints with no targets - nothing to create
		if len(ep.Targets) == 0 {
			p.Log.Debugf("Skipping endpoint with no targets: %s", ep.DNSName)
			continue
		}

		// Process all record types the same way
		records, err := NewDnsRecordsFromEndpoint(ep)
		if err != nil {
			return nil, fmt.Errorf("failed to create records from endpoint %s: %v", ep.DNSName, err)
		}

		p.Log.Debugf("Split endpoint %s into %d record(s)", ep.DNSName, len(records))

		for _, record := range records {
			setID := record.GenerateSetIdentifier()

			e := &endpoint.Endpoint{
				DNSName:          ep.DNSName,
				Targets:          record.GetTarget(),
				RecordType:       ep.RecordType,
				SetIdentifier:    setID,
				RecordTTL:        ep.RecordTTL,
				Labels:           ep.Labels,
				ProviderSpecific: ep.ProviderSpecific,
			}

			if ep.RecordType != endpoint.RecordTypeTXT {
				e.WithLabel("set-identifier", setID)
				p.Log.Debugf("Added set-identifier label to %s record: %s -> %v (SetIdentifier: %s)", ep.RecordType, ep.DNSName, record.GetTarget(), setID)
			} else {
				p.Log.Debugf("Skipping set-identifier label for TXT record: %s -> %v (SetIdentifier: %s)", ep.DNSName, record.GetTarget(), setID)
			}

			adjusted = append(adjusted, e)
		}
	}

	p.Log.Debugf("AdjustEndpoints returning %d endpoint(s)", len(adjusted))
	return adjusted, nil
}

// ApplyChanges applies a set of changes to OPNsense Unbound DNS.
func (p *Provider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	p.Log.Debugf("ApplyChanges called with %d creates, %d updates, %d deletes", len(changes.Create), len(changes.UpdateNew), len(changes.Delete))

	for _, ep := range changes.Delete {
		p.Log.Debugf("Delete request for: %s (%s), SetIdentifier: %s, Targets: %v, Labels: %+v", ep.DNSName, ep.RecordType, ep.SetIdentifier, ep.Targets, ep.Labels)

		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			record, err := NewDnsRecordFromExistingEndpoint(ep)
			if err != nil {
				return fmt.Errorf("failed to create record from endpoint %s: %w", ep.DNSName, err)
			}

			p.Log.Debugf("Deleting host override: %s (%s) with id %s, SetIdentifier: %s", ep.DNSName, ep.RecordType, record.Id, ep.SetIdentifier)
			if err := p.Client.UnboundDeleteHostOverride(ctx, record.Id); err != nil {
				return fmt.Errorf("failed to delete host override %s: %w", ep.DNSName, err)
			}
			p.Log.Infof("Deleted host override: %s (%s) with id %s, SetIdentifier: %s", ep.DNSName, ep.RecordType, record.Id, ep.SetIdentifier)
		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", ep.RecordType, ep.DNSName)
		}
	}

	// UpdateOld and UpdateNew are parallel arrays with matching indices
	for i, n := range changes.UpdateNew {
		o := changes.UpdateOld[i]

		switch n.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			uuid, exists := o.GetProviderSpecificProperty(ProviderSpecificUUID.String())
			if !exists {
				return fmt.Errorf("can not find uuid in old endpoint: %s %s", o.RecordType, o.DNSName)
			}

			record, err := NewDnsRecordFromExistingEndpoint(n)
			if err != nil {
				return fmt.Errorf("failed to create record from endpoint %s: %w", n.DNSName, err)
			}
			record.Id = uuid

			p.Log.Debugf("Updating host override: %s (%s) with id %s", n.DNSName, n.RecordType, record.Id)
			if err := p.Client.UnboundUpdateHostOverride(ctx, record.Id, record.IntoHostOverride()); err != nil {
				return fmt.Errorf("failed to update host override %s: %w", n.DNSName, err)
			}
			p.Log.Infof("Updated host override: %s (%s) with id %s", n.DNSName, n.RecordType, record.Id)

		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", n.RecordType, n.DNSName)
		}
	}

	for _, ep := range changes.Create {
		p.Log.Debugf("Create request for: %s (%s), SetIdentifier: %s, Targets: %v, Labels: %+v", ep.DNSName, ep.RecordType, ep.SetIdentifier, ep.Targets, ep.Labels)

		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			records, err := NewDnsRecordsFromEndpoint(ep)
			if err != nil {
				return fmt.Errorf("failed to create records from endpoint %s: %w", ep.DNSName, err)
			}

			for _, record := range records {
				p.Log.Debugf("Creating host override: %s (%s) -> %s, SetIdentifier: %s", ep.DNSName, ep.RecordType, record.GetTarget(), ep.SetIdentifier)
				if _, err := p.Client.UnboundCreateHostOverride(ctx, record.IntoHostOverride()); err != nil {
					return fmt.Errorf("failed to create host override %s: %w", ep.DNSName, err)
				}
				p.Log.Infof("Created host override: %s (%s) -> %s, SetIdentifier: %s", ep.DNSName, ep.RecordType, record.GetTarget(), ep.SetIdentifier)
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
