package webhook

import (
	"net/http"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"sigs.k8s.io/external-dns/endpoint"
)

func (h *Handler) HandleAdjustEndpointsPost(c *ctx.Context) error {
	if err := h.VerifyHeaders(c); err != nil {
		return err
	}

	body := []*endpoint.Endpoint{}
	if err := c.BindBody(body); err != nil {
		return err
	}

	endpoints, err := h.Provider.AdjustEndpoints(body)
	if err != nil {
		return c.NewHTTPError(http.StatusUnprocessableEntity, err)
	}

	return c.JSON(http.StatusOK, endpoints)
}
