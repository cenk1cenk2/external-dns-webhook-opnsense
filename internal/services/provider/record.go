package provider

import (
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/browningluke/opnsense-go/pkg/unbound"
	"sigs.k8s.io/external-dns/endpoint"
)

type Record struct {
	unbound.SearchHostOverrideItem
	RecordType string
}

func NewRecord(override unbound.SearchHostOverrideItem) *Record {
	return &Record{
		SearchHostOverrideItem: override,
	}
}

func NewRecordFromEndpoint(ep *endpoint.Endpoint) (*Record, error) {
	if slices.Contains([]string{endpoint.RecordTypeA, endpoint.RecordTypeAAAA}, ep.RecordType) {
		return nil, fmt.Errorf("unsupported record type: %s", ep.RecordType)
	}

	record := &Record{
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

	ip := net.ParseIP(record.Server)

	if ip.To4() != nil {
		record.RecordType = endpoint.RecordTypeA
	} else if ip.To16() != nil {
		record.RecordType = endpoint.RecordTypeAAAA
	} else {
		return nil, fmt.Errorf("invalid record type: %s", record.Server)
	}

	return record, nil
}

func NewRecordFromExistingEndpoint(ep *endpoint.Endpoint) (*Record, error) {
	id, exists := ep.GetProviderSpecificProperty(ProviderSpecificUUID.String())
	if !exists {
		return nil, fmt.Errorf("provider specific id not found attached to the endpoint")
	}

	record, err := NewRecordFromEndpoint(ep)
	if err != nil {
		return nil, err
	}

	record.Id = id

	return record, nil
}

func (r *Record) IsEnabled() bool {
	return r.Enabled == "1"
}

func (r *Record) IsDrifted() bool {
	return !r.IsEnabled()
}

func (r *Record) GetFQDN() string {
	return fmt.Sprintf("%s.%s", r.Hostname, r.Domain)
}

func (r *Record) GetOwner() string {
	return r.Description
}

func (r *Record) IsOwnedBy(ownership *Ownership) (bool, error) {
	return ownership.IsOwnedRecord(r)
}

func (r *Record) SetOwnedBy(ownership *Ownership) error {
	owned, err := ownership.ToOwnedRecord(r)
	if err != nil {
		return fmt.Errorf("failed to set ownership: %w", err)
	}

	r.Description = owned

	return nil
}

func (r *Record) IntoHostOverride() *unbound.HostOverride {
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
