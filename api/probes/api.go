package probes

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/interfaces"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/labstack/echo/v4"
)

type Api struct {
	Config ApiConfig
	Echo   *echo.Echo
	log    services.ZapSugaredLogger

	*ApiSvc
}

var _ interfaces.RegisterRoutes = (*Api)(nil)

type ApiConfig struct {
}

type ApiSvc struct {
	Logger    *services.Logger
	Validator *services.Validator

	WebhookApi *api.Api
}

func NewApi(svc *ApiSvc, conf ApiConfig) *Api {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	a := &Api{
		Config: conf,
		Echo:   e,
		ApiSvc: svc,
		log:    svc.Logger.WithCaller(),
	}

	a.SetupMiddleware()
	a.RegisterRoutes(a.Echo.Group(""))

	return a
}

func (a *Api) Start(address string) chan error {
	err := make(chan error)

	go func() {
		err <- a.Echo.Start(address)
	}()

	a.log.Infof("Starting health server at address: %s", (<-a.GetListener()).Addr().String())

	return err
}

func (a *Api) IsReady() chan bool {
	res := make(chan bool, 1)

	<-a.GetListener()

	<-a.WebhookApi.IsReady()

	res <- true

	return res
}

func (a *Api) GetListener() chan net.Listener {
	listener := make(chan net.Listener, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		<-ctx.Done()

		if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
			a.log.Panicf("Listener not ready: %w", err)
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
