package webhook

import (
	"fmt"
	"net/http"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/labstack/echo/v4"
	externaldnsapi "sigs.k8s.io/external-dns/provider/webhook/api"
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
	ExternalDnsAcceptedMedia string = externaldnsapi.MediaTypeFormatAndVersion
)

func (h *Handler) VerifyHeaders(c *ctx.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		if c.Request().Header.Get(echo.HeaderAccept) != ExternalDnsAcceptedMedia {
			return c.NewHTTPError(
				http.StatusNotAcceptable,
				fmt.Errorf(
					"unsupported media type: got %s, must be %s",
					c.Request().Header.Get(echo.HeaderAccept),
					ExternalDnsAcceptedMedia,
				),
			)
		}
	case http.MethodPost:
		if c.Request().Header.Get(echo.HeaderContentType) != ExternalDnsAcceptedMedia {
			return c.NewHTTPError(
				http.StatusUnsupportedMediaType,
				fmt.Errorf(
					"unsupported media type: got %s, must be %s",
					c.Request().Header.Get(echo.HeaderContentType),
					ExternalDnsAcceptedMedia,
				),
			)
		}

		// we will process this as json from now on
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}

	return nil
}
