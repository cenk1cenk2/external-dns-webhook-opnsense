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
			WithSetIdentifier(record.Id)
		if ep == nil {
			return nil, fmt.Errorf("failed to create endpoint for record %s", record.GetFQDN())
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

// ApplyChanges applies a set of changes to OPNsense Unbound DNS.
func (p *Provider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	for _, ep := range changes.Delete {
		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			if ep.SetIdentifier == "" {
				return fmt.Errorf("no opnsense identifier found for endpoint %s", ep.DNSName)
			}

			p.Log.Debugf("Deleting domain override: %s (%s) with id %s", ep.DNSName, ep.RecordType, ep.SetIdentifier)
			if err := p.Client.UnboundDeleteHostOverride(ctx, ep.SetIdentifier); err != nil {
				return fmt.Errorf("failed to delete domain override %s with id %s: %w", ep.DNSName, ep.SetIdentifier, err)
			}
			p.Log.Infof("Deleted domain override: %s (%s) with id %s", ep.DNSName, ep.RecordType, ep.SetIdentifier)
		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", ep.RecordType, ep.DNSName)
		}
	}

	for _, ep := range changes.UpdateNew {
		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			if ep.SetIdentifier == "" {
				return fmt.Errorf("no opnsense identifier found for endpoint %s", ep.DNSName)
			}

			p.Log.Debugf("Deleting old domain override: %s (%s) with id %s", ep.DNSName, ep.RecordType, ep.SetIdentifier)
			if err := p.Client.UnboundDeleteHostOverride(ctx, ep.SetIdentifier); err != nil {
				return fmt.Errorf("failed to delete old domain override %s with id %s: %w", ep.DNSName, ep.SetIdentifier, err)
			}

			records, err := NewDnsRecordsFromEndpoint(ep)
			if err != nil {
				return fmt.Errorf("failed to create records from endpoint %s: %w", ep.DNSName, err)
			}

			for _, record := range records {
				p.Log.Debugf("Creating new domain override: %s (%s) -> %s", ep.DNSName, ep.RecordType, record.GetTarget())
				if _, err := p.Client.UnboundCreateHostOverride(ctx, record.IntoHostOverride()); err != nil {
					return fmt.Errorf("failed to create new domain override %s: %w", ep.DNSName, err)
				}
				p.Log.Infof("Updated domain override: %s (%s) -> %s", ep.DNSName, ep.RecordType, record.GetTarget())
			}

		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", ep.RecordType, ep.DNSName)
		}
	}

	for _, ep := range changes.Create {
		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			records, err := NewDnsRecordsFromEndpoint(ep)
			if err != nil {
				return fmt.Errorf("failed to create records from endpoint %s: %w", ep.DNSName, err)
			}

			for _, record := range records {
				p.Log.Debugf("Creating domain override: %s (%s) -> %s", ep.DNSName, ep.RecordType, record.GetTarget())
				if _, err := p.Client.UnboundCreateHostOverride(ctx, record.IntoHostOverride()); err != nil {
					return fmt.Errorf("failed to create domain override %s: %w", ep.DNSName, err)
				}
				p.Log.Infof("Created domain override: %s (%s) -> %s", ep.DNSName, ep.RecordType, record.GetTarget())
			}

		default:
			p.Log.Warnf("Record type is not supported: %s -> %s", ep.RecordType, ep.DNSName)
		}

	}

	return nil
}

func (p *Provider) AdjustEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error) {
	return endpoints, nil
}

// GetDomainFilter returns the domain filter for this provider.
func (p *Provider) GetDomainFilter() endpoint.DomainFilterInterface {
	return p.DomainFilter
}
