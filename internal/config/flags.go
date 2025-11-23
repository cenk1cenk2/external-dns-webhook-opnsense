package config

import (
	"github.com/urfave/cli/v3"
)

//revive:disable:line-length-limit

func GetFlags(c *Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "log-level",
			Usage: `Log level for the application. enum("debug", "info", "warning", "error", "fatal")`,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("LOG_LEVEL"),
			),
			Required:    false,
			Value:       "info",
			Destination: &c.LogLevel,
		},

		&cli.StringFlag{
			Name:  "log-encoder",
			Usage: `Log encoder format. enum("console", "json")`,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("LOG_ENCODER"),
			),
			Required:    false,
			Value:       "json",
			Destination: &c.LogEncoder,
		},

		&cli.Uint16Flag{
			Name:  "port",
			Usage: "Port on which the server will listen.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PORT"),
			),
			Required:    false,
			Value:       8888,
			Destination: &c.Port,
		},

		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "The application will not make any changes to the OPNsense DNS records, only log the intended actions.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("DRY_RUN"),
			),
			Required:    false,
			Value:       false,
			Destination: &c.Api.OpnsenseClient.DryRun,
		},

		&cli.StringFlag{
			Name:  "opnsense-uri",
			Usage: "The base URI of the OPNsense API endpoint.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_URI"),
			),
			Required:    true,
			Destination: &c.Api.OpnsenseClient.Options.Uri,
		},

		&cli.StringFlag{
			Name:  "opnsense-api-key",
			Usage: "The API key for authenticating with the OPNsense API.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_API_KEY"),
			),
			Required:    true,
			Destination: &c.Api.OpnsenseClient.Options.APIKey,
		},

		&cli.StringFlag{
			Name:  "opnsense-api-secret",
			Usage: "The API secret for authenticating with the OPNsense API.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_API_SECRET"),
			),
			Required:    true,
			Destination: &c.Api.OpnsenseClient.Options.APISecret,
		},

		&cli.BoolFlag{
			Name:  "opnsense-allow-insecure",
			Usage: "Allow insecure TLS connections to the OPNsense API.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_ALLOW_INSECURE"),
			),
			Required:    false,
			Value:       false,
			Destination: &c.Api.OpnsenseClient.Options.AllowInsecure,
		},

		&cli.Int64Flag{
			Name:  "opnsense-max-backoff",
			Usage: "Maximum backoff time in seconds for retrying OPNsense API requests.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_MAX_BACKOFF"),
			),
			Required:    false,
			Value:       120,
			Destination: &c.Api.OpnsenseClient.Options.MaxBackoff,
		},

		&cli.Int64Flag{
			Name:  "opnsense-min-backoff",
			Usage: "Minimum backoff time in seconds for retrying OPNsense API requests.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_MIN_BACKOFF"),
			),
			Required:    false,
			Value:       120,
			Destination: &c.Api.OpnsenseClient.Options.MinBackoff,
		},

		&cli.Int64Flag{
			Name:  "opnsense-max-retries",
			Usage: "Maximum retries for OPNsense API requests.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_MAX_RETRIES"),
			),
			Required:    false,
			Value:       120,
			Destination: &c.Api.OpnsenseClient.Options.MaxRetries,
		},

		&cli.StringSliceFlag{
			Name:  "domain-include-filter",
			Usage: "List of domain include filters. Only domains matching these filters will be managed. Can be specified multiple times.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("DOMAIN_INCLUDE_FILTER"),
			),
			Required:    false,
			Destination: &c.Api.Provider.DomainIncludeFilter,
		},

		&cli.StringSliceFlag{
			Name:  "domain-exclude-filter",
			Usage: "List of domain exclude filters. Domains matching these filters will be ignored. Can be specified multiple times.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("DOMAIN_EXCLUDE_FILTER"),
			),
			Required:    false,
			Destination: &c.Api.Provider.DomainExcludeFilter,
		},
	}
}
