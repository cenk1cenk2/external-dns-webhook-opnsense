package webhook

import (
	"net/http"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HandleNegotiateGet(c *ctx.Context) error {
	if err := h.VerifyHeaders(c); err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, ExternalDnsAcceptedMedia)

	return c.JSON(http.StatusOK, h.Provider.GetDomainFilter())
}
