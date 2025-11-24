package config

import (
	"github.com/urfave/cli/v3"
)

//revive:disable:line-length-limit

func BindFlags(c *Config) []cli.Flag {
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
			Destination: &c.OpnsenseClient.DryRun,
		},

		&cli.StringFlag{
			Name:  "opnsense-url",
			Usage: "The base URI of the OPNsense API endpoint.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_URL"),
			),
			Required:    true,
			Destination: &c.OpnsenseClient.Options.Uri,
		},

		&cli.StringFlag{
			Name:  "opnsense-api-key",
			Usage: "The API key for authenticating with the OPNsense API.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_API_KEY"),
			),
			Required:    true,
			Destination: &c.OpnsenseClient.Options.APIKey,
		},

		&cli.StringFlag{
			Name:  "opnsense-api-secret",
			Usage: "The API secret for authenticating with the OPNsense API.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_API_SECRET"),
			),
			Required:    true,
			Destination: &c.OpnsenseClient.Options.APISecret,
		},

		&cli.BoolFlag{
			Name:  "opnsense-allow-insecure",
			Usage: "Allow insecure TLS connections to the OPNsense API.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_ALLOW_INSECURE"),
			),
			Required:    false,
			Value:       false,
			Destination: &c.OpnsenseClient.Options.AllowInsecure,
		},

		&cli.Int64Flag{
			Name:  "opnsense-max-backoff",
			Usage: "Maximum backoff time in seconds for retrying OPNsense API requests.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_MAX_BACKOFF"),
			),
			Required:    false,
			Value:       120,
			Destination: &c.OpnsenseClient.Options.MaxBackoff,
		},

		&cli.Int64Flag{
			Name:  "opnsense-min-backoff",
			Usage: "Minimum backoff time in seconds for retrying OPNsense API requests.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_MIN_BACKOFF"),
			),
			Required:    false,
			Value:       120,
			Destination: &c.OpnsenseClient.Options.MinBackoff,
		},

		&cli.Int64Flag{
			Name:  "opnsense-max-retries",
			Usage: "Maximum retries for OPNsense API requests.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("OPNSENSE_MAX_RETRIES"),
			),
			Required:    false,
			Value:       120,
			Destination: &c.OpnsenseClient.Options.MaxRetries,
		},

		// match with upstream: https://github.com/kubernetes-sigs/external-dns/blob/master/docs/flags.md

		&cli.StringSliceFlag{
			Name:  "domain-filter",
			Usage: "List of domain include filters.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("DOMAIN_FILTER"),
			),
			Required:    false,
			Destination: &c.Provider.DomainFilter.DomainFilter,
		},

		&cli.StringSliceFlag{
			Name:  "exclude-domains",
			Usage: "List of domain exclude filters.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("EXCLUDE_DOMAINS"),
			),
			Required:    false,
			Destination: &c.Provider.DomainFilter.ExcludeDomains,
		},

		&cli.StringFlag{
			Name:  "regex-domain-filter",
			Usage: "List of domain exclude filters in regex form.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("REGEX_DOMAIN_FILTER"),
			),
			Required:    false,
			Destination: &c.Provider.DomainFilter.RegexDomainFilter,
		},

		&cli.StringFlag{
			Name:  "regex-domain-exclusion",
			Usage: "List of domain exclude filters in regex form.",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("REGEX_DOMAIN_EXCLUSION"),
			),
			Required:    false,
			Destination: &c.Provider.DomainFilter.RegexDomainExclusion,
		},

		&cli.StringFlag{
			Name:  "txt-prefix",
			Usage: "Prefix for TXT records used for ownership verification. format(template)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("TXT_PREFIX"),
			),
			Required:    false,
			Value:       "{{ .Type | lower }}-",
			Destination: &c.Provider.TxtPrefix,
		},

		&cli.StringFlag{
			Name:  "txt-suffix",
			Usage: "Suffix for TXT records used for ownership verification. format(template)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("TXT_SUFFIX"),
			),
			Required:    false,
			Value:       "",
			Destination: &c.Provider.TxtSuffix,
		},
	}
}
