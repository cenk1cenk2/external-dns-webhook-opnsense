package probes

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
)

type Handler struct {
	*HandlerSvc
}

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
