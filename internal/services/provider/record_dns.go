package provider

import (
	"fmt"
	"slices"
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
	if slices.Contains([]string{endpoint.RecordTypeA, endpoint.RecordTypeAAAA}, ep.RecordType) {
		return nil, fmt.Errorf("unsupported record type: %s", ep.RecordType)
	}

	record := &DnsRecord{
		SearchHostOverrideItem: unbound.SearchHostOverrideItem{
			Enabled: "1",
		},
	}

	dnsname := strings.SplitN(ep.DNSName, ".", 2)
	record.Hostname = dnsname[0]
	record.Domain = dnsname[1]
	record.Server = ep.Targets[0]

	if record.Hostname == "*" {
		return nil, fmt.Errorf("wildcard hostnames are not supported in opnsense: %s", ep.DNSName)
	}

	return record, nil
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
	return fmt.Sprintf("%s.%s", r.Hostname, r.Domain)
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
	}
}
