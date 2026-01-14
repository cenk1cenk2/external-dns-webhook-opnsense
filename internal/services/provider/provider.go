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
		return nil, fmt.Errorf("failed to query for domain overrides: %w", err)
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

		if record.IsDrifted() {
			p.Log.Debugf("Record has drifted, will fix at next adjust: %s", record.GetFQDN())

			ep.WithProviderSpecific(
				ProviderSpecificDrifted.String(),
				"true",
			)
		}

		p.Log.Debugf("Endpoint processed: %+v", ep)

		endpoints = append(endpoints, ep)
	}

	return endpoints, nil
}

// ApplyChanges applies a set of changes to OPNsense Unbound DNS.
func (p *Provider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	for _, ep := range changes.Delete {
		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			record, err := NewDnsRecordFromExistingEndpoint(ep)
			if err != nil {
				return fmt.Errorf("failed to create record from endpoint %s: %w", ep.DNSName, err)
			}

			p.Log.Debugf("Deleting domain override: %s (%s) with id %s", ep.DNSName, ep.RecordType, record)
			if err := p.Client.UnboundDeleteHostOverride(ctx, record.Id); err != nil {
				return fmt.Errorf("failed to delete domain override %s: %w", ep.DNSName, err)
			}
			p.Log.Infof("Deleted domain override: %s (%s) with id %s", ep.DNSName, ep.RecordType, record.Id)
		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", ep.RecordType, ep.DNSName)
		}
	}

	for _, ep := range changes.UpdateNew {
		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			// TODO: need to find the provider specific property probably from the old endpoint probably

			record, err := NewDnsRecordFromExistingEndpoint(ep)
			if err != nil {
				return fmt.Errorf("failed to create record from endpoint %s: %w", ep.DNSName, err)
			}

			p.Log.Debugf("Updating domain override: %s (%s) with id %s", ep.DNSName, ep.RecordType, record.Id)
			if err := p.Client.UnboundUpdateHostOverride(ctx, record.Id, record.IntoHostOverride()); err != nil {
				return fmt.Errorf("failed to update domain override %s: %w", ep.DNSName, err)
			}
			p.Log.Infof("Updated domain override: %s (%s) with id %s", ep.DNSName, ep.RecordType, record.Id)

		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", ep.RecordType, ep.DNSName)
		}
	}

	for _, ep := range changes.Create {
		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			record, err := NewDnsRecordFromEndpoint(ep)
			if err != nil {
				return fmt.Errorf("failed to create record from endpoint %s: %w", ep.DNSName, err)
			}

			p.Log.Debugf("Creating domain override: %s (%s)", ep.DNSName, ep.RecordType)
			if _, err := p.Client.UnboundCreateHostOverride(ctx, record.IntoHostOverride()); err != nil {
				return fmt.Errorf("failed to create domain override %s: %w", ep.DNSName, err)
			}
			p.Log.Infof("Created domain override: %s (%s)", ep.DNSName, ep.RecordType)

		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", ep.RecordType, ep.DNSName)
		}

	}

	return nil
}

func (p *Provider) AdjustEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error) {
	for _, ep := range endpoints {
		_, drifted := ep.GetProviderSpecificProperty(ProviderSpecificDrifted.String())

		if !drifted {
			continue
		}

		p.Log.Warnf("Endpoint is drifted, adjusting accordingly: %s", ep.DNSName)

		record, err := NewDnsRecordFromExistingEndpoint(ep)
		if err != nil {
			return nil, fmt.Errorf("failed to create record from endpoint %s: %w", ep.DNSName, err)
		}

		p.Log.Debugf("Updating drifted endpoint: %s (%s)", ep.DNSName, ep.RecordType)
		if err := p.Client.UnboundUpdateHostOverride(context.Background(), record.Id, record.IntoHostOverride()); err != nil {
			return nil, fmt.Errorf("failed to update drifted endpoint %s: %w", ep.DNSName, err)
		}
	}

	return endpoints, nil
}

// GetDomainFilter returns the domain filter for this provider.
func (p *Provider) GetDomainFilter() endpoint.DomainFilterInterface {
	return p.DomainFilter
}
