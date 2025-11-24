package webhook

import (
	"fmt"
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

const (
	ExternalDnsAcceptedMedia string = "application/external.dns.webhook+json;version=1"
)

func (h *Handler) VerifyHeaders(c *ctx.Context) error {
	if !strings.Contains(c.Request().Header.Get(echo.HeaderAccept), ExternalDnsAcceptedMedia) {
		return c.NewHTTPError(http.StatusNotAcceptable, fmt.Errorf("unsupported media type, must be %s", ExternalDnsAcceptedMedia))
	}

	return nil
}
