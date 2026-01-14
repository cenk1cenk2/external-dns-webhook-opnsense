package provider

import (
	"fmt"
	"strings"

	"github.com/browningluke/opnsense-go/pkg/unbound"
	"sigs.k8s.io/external-dns/endpoint"
)

type DnsRecord struct {
	unbound.SearchHostOverrideItem
}

func NewDnsRecord(override unbound.SearchHostOverrideItem) *DnsRecord {
	return &DnsRecord{
		SearchHostOverrideItem: override,
	}
}

func NewDnsRecordFromEndpoint(ep *endpoint.Endpoint) (*DnsRecord, error) {
	record := &DnsRecord{
		SearchHostOverrideItem: unbound.SearchHostOverrideItem{
			Enabled: "1",
			Type:    ep.RecordType,
		},
	}

	switch ep.RecordType {
	case endpoint.RecordTypeA, endpoint.RecordTypeAAAA:
		dnsname := strings.SplitN(ep.DNSName, ".", 2)
		if len(dnsname) != 2 {
			return nil, fmt.Errorf("invalid dns name: %s", ep.DNSName)
		}
		record.Hostname = dnsname[0]
		record.Domain = dnsname[1]
		if len(ep.Targets) == 0 {
			return nil, fmt.Errorf("no targets found for endpoint: %s", ep.DNSName)
		} else if len(ep.Targets) > 1 {
			return nil, fmt.Errorf("multiple targets can not be handled: %s", ep.DNSName)
		}
		record.Server = ep.Targets[0]

		if record.Hostname == "*" {
			return nil, fmt.Errorf("wildcard hostnames are not supported in opnsense: %s", ep.DNSName)
		}

		return record, nil
	case endpoint.RecordTypeTXT:
		record.Domain = ep.DNSName
		if len(ep.Targets) == 0 {
			return nil, fmt.Errorf("no targets found for endpoint: %s", ep.DNSName)
		} else if len(ep.Targets) > 1 {
			return nil, fmt.Errorf("multiple targets can not be handled: %s", ep.DNSName)
		}
		record.TxtData = ep.Targets[0]

		if record.Hostname == "*" {
			return nil, fmt.Errorf("wildcard hostnames are not supported in opnsense: %s", ep.DNSName)
		}

		return record, nil
	}

	return nil, fmt.Errorf("unsupported record type: %s", ep.RecordType)
}

func NewDnsRecordFromExistingEndpoint(ep *endpoint.Endpoint) (*DnsRecord, error) {
	id, exists := ep.GetProviderSpecificProperty(ProviderSpecificUUID.String())
	if !exists {
		return nil, fmt.Errorf("provider specific id not found attached to the endpoint")
	}

	record, err := NewDnsRecordFromEndpoint(ep)
	if err != nil {
		return nil, err
	}

	record.Id = id

	return record, nil
}

func (r *DnsRecord) IsEnabled() bool {
	return r.Enabled == "1"
}

func (r *DnsRecord) IsDrifted() bool {
	return !r.IsEnabled()
}

func (r *DnsRecord) GetFQDN() string {
	// TXT records store the full FQDN in Domain field with empty Hostname
	if r.Type == endpoint.RecordTypeTXT {
		return r.Domain
	}
	return fmt.Sprintf("%s.%s", r.Hostname, r.Domain)
}

func (r *DnsRecord) GetTarget() []string {
	switch r.Type {
	case endpoint.RecordTypeA, endpoint.RecordTypeAAAA:
		return []string{r.Server}
	case endpoint.RecordTypeTXT:
		return []string{r.TxtData}
	default:
		return []string{}
	}
}

func (r *DnsRecord) IntoHostOverride() *unbound.HostOverride {
	return &unbound.HostOverride{
		Enabled:     r.Enabled,
		Hostname:    r.Hostname,
		Domain:      r.Domain,
		Type:        r.Type,
		Server:      r.Server,
		MXPriority:  r.MXPriority,
		MXDomain:    r.MXDomain,
		Description: r.Description,
		TxtData:     r.TxtData,
	}
}
