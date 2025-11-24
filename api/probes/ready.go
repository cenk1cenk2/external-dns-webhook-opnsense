package probes

import (
	"fmt"
	"net/http"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
)

// @Tags		Probes
// @Summary	Returns the ready status of the service.
// @Produce	json
// @Success	200	{string}	string
// @Router  /readyz [get]
func (h *Handler) HandleReadyGet(c *ctx.Context) error {
	ready := <-h.IsReady()

	if !ready {
		return c.NewHTTPError(http.StatusServiceUnavailable, fmt.Errorf("service is not ready."))
	}

	return c.NoContent(http.StatusOK)
}
