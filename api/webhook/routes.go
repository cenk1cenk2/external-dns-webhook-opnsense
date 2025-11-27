package webhook

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/labstack/echo/v4"
)

func (h *Handler) RegisterRoutes(r *echo.Group) {
	g := r.Group("")

	// as per: https://github.com/kubernetes-sigs/external-dns/blob/0c39b6eb4902ca80cf18f16305e0cb54619daa23/docs/tutorials/webhook-provider.md

	g.GET("/", ctx.With(
		h.HandleNegotiateGet,
		h.Log,
	))

	g.GET("/records", ctx.With(
		h.HandleRecordsGet,
		h.Log,
	))

	g.POST("/records", ctx.With(
		h.HandleRecordsPost,
		h.Log,
	))

	g.POST("/adjustendpoints", ctx.With(
		h.HandleAdjustEndpointsPost,
		h.Log,
	))
}
