package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/interfaces"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/opnsense"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/labstack/echo/v5"
)

type Api struct {
	Config   ApiConfig
	Echo     *echo.Echo
	log      services.ZapSugaredLogger
	listener net.Listener
	server   *http.Server

	*ApiSvc
}

var _ interfaces.RegisterRoutes = (*Api)(nil)

type ApiConfig struct {
}

type ApiSvc struct {
	Logger    *services.Logger
	Validator *services.Validator

	Provider       *provider.Provider
	OpnsenseClient *opnsense.Client
}

func NewApi(svc *ApiSvc, conf ApiConfig) *Api {
	e := echo.New()

	a := &Api{
		Config: conf,
		Echo:   e,
		ApiSvc: svc,
		log:    svc.Logger.WithCaller(),
		server: &http.Server{},
	}

	a.SetupMiddleware()
	a.RegisterRoutes(a.Echo.Group(""))

	return a
}

func (a *Api) Start(address string) chan error {
	errCh := make(chan error)
	isListenerReady := make(chan struct{})

	go func() {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			errCh <- err

			return
		}

		a.listener = listener
		a.server.Handler = a.Echo
		close(isListenerReady)

		a.log.Infof("Starting server at address: %s", listener.Addr().String())

		errCh <- a.server.Serve(listener)
	}()

	<-isListenerReady

	return errCh
}

func (a *Api) IsReady() chan bool {
	res := make(chan bool, 1)

	<-a.GetListener()

	if err := a.OpnsenseClient.CheckUnboundService(context.Background()); err != nil {
		a.log.Errorf("Unbound service is not running: %v", err)

		res <- false

		return res
	}

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

	for a.listener == nil {
		time.Sleep(10 * time.Millisecond)
	}

	listener <- a.listener

	return listener
}

func (a *Api) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.server.Shutdown(ctx)
}
