package probes

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/labstack/echo/v4"
)

func (h *Handler) RegisterRoutes(r *echo.Group) *Handler {
	g := r.Group("")

	g.GET("/healthz", ctx.With(
		h.HandleHealthGet,
		h.Log,
	))
	g.GET("/readyz", ctx.With(
		h.HandleReadyGet,
		h.Log,
	))

	return h
}
