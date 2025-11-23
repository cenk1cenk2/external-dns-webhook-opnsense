package provider

import (
	"sigs.k8s.io/external-dns/endpoint"
)

type DomainFilter struct {
}

var _ endpoint.DomainFilterInterface = (*DomainFilter)(nil)

func NewDomainFilter() *DomainFilter {
	return &DomainFilter{}
}

func (d *DomainFilter) Match(domain string) bool {
	// TODO: implement me, this should include the regex filtering logic
	return true
}
