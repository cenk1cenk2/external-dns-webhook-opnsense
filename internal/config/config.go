package config

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api"
)

type Config struct {
	LogLevel   string
	LogEncoder string

	Port uint16

	Api api.ApiConfig
}

func NewConfig() *Config {
	return &Config{}
}
