package api

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/opnsense"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/labstack/echo/v4"
)

type Api struct {
	Config ApiConfig
	Echo   *echo.Echo
	*ApiSvc
}

type ApiConfig struct {
}

type ApiSvc struct {
	Log       *services.Logger
	Validator *services.Validator

	Provider       *provider.Provider
	OpnsenseClient *opnsense.Client
}

func NewApi(svc *ApiSvc, conf ApiConfig) *Api {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	a := &Api{
		Config: conf,
		Echo:   e,
		ApiSvc: svc,
	}

	a.SetupMiddleware()
	a.RegisterRoutes()

	return a
}

func (a *Api) Start(address string) chan error {
	log := a.Log.WithCaller()

	err := make(chan error)

	go func() {
		err <- a.Echo.Start(address)
	}()

	log.Infof("Starting server at address: %s", (<-a.GetListener()).Addr().String())

	return err
}

func (a *Api) IsReady() chan bool {
	res := make(chan bool, 1)

	<-a.GetListener()

	res <- true

	return res
}

func (a *Api) GetListener() chan net.Listener {
	log := a.Log.WithCaller()
	listener := make(chan net.Listener, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		<-ctx.Done()

		if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
			log.Panicf("Listener not ready: %w", err)
		}
	}()

	for a.Echo.Listener == nil {
		time.Sleep(0)
	}

	listener <- a.Echo.Listener

	return listener
}

func (a *Api) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.Echo.Shutdown(ctx)
}
