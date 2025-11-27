package probes

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/labstack/echo/v4"
)

func (a *Api) RegisterRoutes(group *echo.Group) {
	NewHandler(&HandlerSvc{
		Log:     a.Logger,
		IsReady: a.IsReady,
	}).
		RegisterRoutes(group)
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/healthz", ctx.With(
		h.HandleHealthGet,
		h.Log,
	))
	g.GET("/readyz", ctx.With(
		h.HandleReadyGet,
		h.Log,
	))
}
