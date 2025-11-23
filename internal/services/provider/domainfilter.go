package provider

import (
	"regexp"

	"sigs.k8s.io/external-dns/endpoint"
)

type DomainFilter struct {
	*endpoint.DomainFilter
}

type DomainFilterConfig struct {
	DomainFilter         []string
	ExcludeDomains       []string
	RegexDomainFilter    string
	RegexDomainExclusion string
}

var _ endpoint.DomainFilterInterface = (*DomainFilter)(nil)

func NewDomainFilter(conf DomainFilterConfig) *DomainFilter {
	if len(conf.RegexDomainFilter) > 0 {
		return &DomainFilter{
			DomainFilter: endpoint.NewRegexDomainFilter(
				regexp.MustCompile(conf.RegexDomainFilter),
				regexp.MustCompile(conf.RegexDomainExclusion),
			),
		}

	}

	return &DomainFilter{
		DomainFilter: endpoint.NewDomainFilterWithExclusions(conf.DomainFilter, conf.ExcludeDomains),
	}
}
