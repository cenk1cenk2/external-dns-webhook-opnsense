package webhook

import (
	"net/http"
	"strings"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	*HandlerSvc
}

type HandlerSvc struct {
	Log      *services.Logger
	Provider *provider.Provider
}

func NewHandler(svc *HandlerSvc) *Handler {
	h := &Handler{
		HandlerSvc: svc,
	}

	return h
}

func (h *Handler) VerifyHeaders(c *ctx.Context) error {
	accept := c.Request().Header.Get(echo.HeaderAccept)
	if !strings.Contains(accept, "application/external.dns.webhook+json;version=1") {
		c.Log.Debugf("does not containt: %s", accept)

		return c.NoContent(http.StatusNotAcceptable)
	} else {
		c.Log.Debugf("Content-Type: %s", accept)
	}

	content := c.Request().Header.Get(echo.HeaderContentType)
	if !strings.Contains(content, echo.MIMEApplicationJSON) {
		c.Log.Debugf("does not containt: %s", content)
		return c.NoContent(http.StatusUnsupportedMediaType)
	} else {
		c.Log.Debugf("Content-Type: %s", content)
	}

	return nil
}
