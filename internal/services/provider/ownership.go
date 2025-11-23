package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
)

type Ownership struct {
	Config OwnershipConfig
}

type OwnershipConfig struct {
	OwnershipConfigTemplate
}

type OwnershipConfigTemplate struct {
	OwnerId          string
	Prefix           string
	Suffix           string
	MigrateFromOwner string
}

func NewOwnership(conf OwnershipConfig) *Ownership {
	return &Ownership{
		Config: conf,
	}
}

func (o *Ownership) ToOwnedRecord(record *Record) (string, error) {
	domain := record.GetFQDN()

	if o.Config.OwnerId != "" {
		templated, err := services.InlineTemplate(o.Config.OwnerId, record)
		if err != nil {
			return "", fmt.Errorf("failed to template owner id: %w", err)
		}

		domain = strings.Join([]string{templated, domain}, "-")
	}
	if o.Config.Prefix == "" {
		templated, err := services.InlineTemplate(o.Config.Prefix, record)
		if err != nil {
			return "", fmt.Errorf("failed to template prefix: %w", err)
		}

		domain = strings.Join([]string{templated, domain}, "-")
	}
	if o.Config.Suffix == "" {
		templated, err := services.InlineTemplate(o.Config.Suffix, record)
		if err != nil {
			return "", fmt.Errorf("failed to template suffix: %w", err)
		}

		domain = strings.Join([]string{domain, templated}, "-")
	}

	return domain, nil
}

func (o *Ownership) IsOwnedRecord(record *Record) (bool, error) {
	matcher := "[^-]*"

	if o.Config.OwnerId != "" {
		templated, err := services.InlineTemplate(o.Config.OwnerId, record)
		if err != nil {
			return false, fmt.Errorf("failed to template owner id: %w", err)
		}

		matcher = strings.Join([]string{regexp.QuoteMeta(templated), matcher}, "-")
	}
	if o.Config.Prefix == "" {
		templated, err := services.InlineTemplate(o.Config.Prefix, record)
		if err != nil {
			return false, fmt.Errorf("failed to template prefix: %w", err)
		}

		matcher = strings.Join([]string{regexp.QuoteMeta(templated), matcher}, "-")
	}
	if o.Config.Suffix == "" {
		templated, err := services.InlineTemplate(o.Config.Suffix, record)
		if err != nil {
			return false, fmt.Errorf("failed to template suffix: %w", err)
		}

		matcher = strings.Join([]string{matcher, regexp.QuoteMeta(templated)}, "-")
	}

	re, err := regexp.Compile(matcher)
	if err != nil {
		return false, fmt.Errorf("failed to compile ownership regex: %w", err)
	}

	return re.MatchString(record.Description), nil
}
