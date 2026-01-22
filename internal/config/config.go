package config

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/probes"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/opnsense"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
)

type Config struct {
	LogLevel   string
	LogEncoder string

	Port       uint16
	HealthPort uint16

	Api    api.ApiConfig
	Probes probes.ApiConfig

	OpnsenseClient opnsense.ClientConfig
	Provider       provider.ProviderConfig
}

func NewConfig() *Config {
	return &Config{}
}
