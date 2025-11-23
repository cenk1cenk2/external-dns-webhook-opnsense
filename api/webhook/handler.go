package webhook

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
)

type Handler struct {
	*HandlerSvc
}

type HandlerSvc struct {
	Log *services.Logger
}

func NewHandler(svc *HandlerSvc) *Handler {
	h := &Handler{
		HandlerSvc: svc,
	}

	return h
}
