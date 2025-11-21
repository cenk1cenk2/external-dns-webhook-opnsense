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
	}
}
