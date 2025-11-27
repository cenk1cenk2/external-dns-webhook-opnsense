package probes

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/interfaces"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
)

type Handler struct {
	*HandlerSvc
}

var _ interfaces.RegisterRoutes = (*Handler)(nil)

type IsReadyFunc = func() chan bool

type HandlerSvc struct {
	Log     *services.Logger
	IsReady IsReadyFunc
}

func NewHandler(svc *HandlerSvc) *Handler {
	h := &Handler{
		HandlerSvc: svc,
	}

	return h
}
