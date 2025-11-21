package probes

import (
	"net/http"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
)

// @Tags		Probes
// @Summary	Returns the health status of the service.
// @Produce	json
// @Success	200	{string}	string
// @Router	/healthz [get]
func (h *Handler) HandleHealthGet(c *ctx.Context) error {
	return c.JSON(http.StatusOK, "")
}
