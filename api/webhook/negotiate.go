package webhook

import (
	"net/http"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
)

func (h *Handler) HandleNegotiateGet(c *ctx.Context) error {
	if err := h.VerifyHeaders(c); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, h.Provider.GetDomainFilter())
}
