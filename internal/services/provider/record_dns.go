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

// NewDnsRecordsFromEndpoint converts an external-dns endpoint into one or more OPNsense DNS records.
// Multiple records are created when the endpoint has multiple targets (for A/AAAA/TXT records).
func NewDnsRecordsFromEndpoint(ep *endpoint.Endpoint) ([]*DnsRecord, error) {
	if len(ep.Targets) == 0 {
		return nil, fmt.Errorf("no targets found for endpoint: %s", ep.DNSName)
	}

	records := make([]*DnsRecord, 0, len(ep.Targets))

	description := ""
	if desc, exists := ep.GetProviderSpecificProperty(ProviderSpecificDescription.String()); exists {
		description = desc
	}

	switch ep.RecordType {
	case endpoint.RecordTypeA, endpoint.RecordTypeAAAA:
		dnsname := strings.SplitN(ep.DNSName, ".", 2)
		if len(dnsname) != 2 {
			return nil, fmt.Errorf("invalid dns name: %s", ep.DNSName)
		}

		hostname := dnsname[0]
		domain := dnsname[1]

		if hostname == "*" {
			return nil, fmt.Errorf("wildcard hostnames are not supported in opnsense: %s", ep.DNSName)
		}

		for _, target := range ep.Targets {
			record := &DnsRecord{
				SearchHostOverrideItem: unbound.SearchHostOverrideItem{
					Enabled:     "1",
					Id:          ep.SetIdentifier,
					Type:        ep.RecordType,
					Hostname:    hostname,
					Domain:      domain,
					Server:      target,
					Description: description,
				},
			}
			records = append(records, record)
		}

		return records, nil

	case endpoint.RecordTypeTXT:
		if ep.DNSName == "*" {
			return nil, fmt.Errorf("wildcard hostnames are not supported in opnsense: %s", ep.DNSName)
		}

		for _, target := range ep.Targets {
			record := &DnsRecord{
				SearchHostOverrideItem: unbound.SearchHostOverrideItem{
					Enabled:     "1",
					Id:          ep.SetIdentifier,
					Type:        ep.RecordType,
					Domain:      ep.DNSName,
					TxtData:     target,
					Description: description,
				},
			}
			records = append(records, record)
		}

		return records, nil
	}

	return nil, fmt.Errorf("unsupported record type: %s", ep.RecordType)
}

func (r *DnsRecord) IsEnabled() bool {
	return r.Enabled == "1"
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
