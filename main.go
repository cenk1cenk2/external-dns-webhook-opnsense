package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/config"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/opnsense"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/urfave/cli/v3"
)

func main() {
	conf := config.NewConfig()

	cmd := &cli.Command{
		Name:    "external-dns-webhook-opnsense",
		Version: VERSION,
		Flags:   config.GetFlags(conf),
		Action: func(_ context.Context, cmd *cli.Command) error {
			logger, err := services.NewLogger(&services.LoggerConfig{
				Level:   conf.LogLevel,
				Encoder: services.LogEncoder(conf.LogEncoder),
			})
			if err != nil {
				return err
			}
			log := logger.WithCaller()

			validator := services.NewValidator()

			if err := validator.Validate(conf); err != nil {
				return err
			}

			// TODO: add a check for opnsense client to do a health call

			client, err := opnsense.NewClient(
				&opnsense.OpnsenseClientSvc{
					Logger: logger,
				},
				conf.OpnsenseClient,
			)
			if err != nil {
				return fmt.Errorf("failed to create opnsense client: %w", err)
			}

			provider, err := provider.NewProvider(
				&provider.ProviderSvc{
					Client: client,
					Logger: logger,
				},
				conf.Provider,
			)
			if err != nil {
				return fmt.Errorf("failed to create provider: %w", err)
			}

			a := api.NewApi(&api.ApiSvc{
				Log:            logger,
				Validator:      validator,
				OpnsenseClient: client,
				Provider:       provider,
			}, conf.Api)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			go func() {
				if err := <-a.Start(fmt.Sprintf(":%d", conf.Port)); err != nil && errors.Is(err, http.ErrServerClosed) {
					log.Warnf("Shutting down the server.")
				} else if err != nil {
					log.Panicf("Failed to start the server: %w", err)
				}
			}()

			<-ctx.Done()
			if err := a.Shutdown(); err != nil {
				log.Warnln(err)

				return err
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err.Error())

		os.Exit(1)
	}
}
