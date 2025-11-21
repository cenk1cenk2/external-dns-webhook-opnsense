package fixtures

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/config"
)

func NewTestConfig() *config.Config {
	return &config.Config{
		LogLevel:   "debug",
		LogEncoder: "console",
		Port:       8888,
		Api:        api.ApiConfig{},
	}
}
