package provider

import (
	"encoding/json"
	"errors"
	"fmt"

	"sigs.k8s.io/external-dns/endpoint"
)

type OwnershipRecord struct {
	Name          string `json:"name"`
	Data          string `json:"data"`
	SetIdentifier string `json:"setIdentifier,omitempty"`
}

var (
	ErrNotOwnershipRecord = errors.New("not an ownership record")
)

func NewOwnershipRecordFromEndpoint(ep *endpoint.Endpoint) (*OwnershipRecord, error) {
	// we replicate the behavior of: https://github.com/kubernetes-sigs/external-dns/blob/master/registry/txt.go#L200
	if ep.RecordType != endpoint.RecordTypeTXT {
		return nil, fmt.Errorf("endpoint is not a TXT record")
	} else if len(ep.Targets) == 0 {
		return nil, fmt.Errorf("endpoint has no targets")
	} else if len(ep.Targets) > 1 {
		return nil, fmt.Errorf("endpoint has multiple targets")
	}
	if _, err := endpoint.NewLabelsFromStringPlain(ep.Targets[0]); errors.Is(err, endpoint.ErrInvalidHeritage) {
		// then this is not a txt ownership record
		return nil, ErrNotOwnershipRecord
	}

	return &OwnershipRecord{
		Name:          ep.DNSName,
		Data:          ep.Targets[0],
		SetIdentifier: ep.SetIdentifier,
	}, nil
}

func NewOwnershipRecordFromDnsRecord(record *DnsRecord) (*OwnershipRecord, error) {
	var ownership *OwnershipRecord

	if err := json.Unmarshal([]byte(record.Description), &ownership); err != nil {
		// this might have no ownership info
		return nil, ErrNotOwnershipRecord
	}

	return ownership, nil
}

func (o *OwnershipRecord) IntoEndpoint() (*endpoint.Endpoint, error) {
	return &endpoint.Endpoint{
		RecordType:    endpoint.RecordTypeTXT,
		DNSName:       o.Name,
		Targets:       []string{o.Data},
		SetIdentifier: o.SetIdentifier,
	}, nil
}

func (o *OwnershipRecord) SetOwnedByForDnsRecord(record *DnsRecord) error {
	ownership, err := o.GetOwnedBy()
	if err != nil {
		return fmt.Errorf("failed to get ownership info: %w", err)
	}

	record.Description = ownership

	return nil
}

func (o *OwnershipRecord) GetOwnedBy() (string, error) {
	ownership, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ownership info: %w", err)
	}

	return string(ownership), nil
}
