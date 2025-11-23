package webhook

import (
	"net/http"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"sigs.k8s.io/external-dns/plan"
)

func (h *Handler) HandleRecordsGet(c *ctx.Context) error {
	if err := h.VerifyHeaders(c); err != nil {
		return err
	}

	endpoints, err := h.Provider.Records(c.Request().Context())
	if err != nil {
		return c.NewHTTPError(http.StatusUnprocessableEntity, err)
	}

	return c.JSON(http.StatusOK, endpoints)
}

func (h *Handler) HandleRecordsPost(c *ctx.Context) error {
	if err := h.VerifyHeaders(c); err != nil {
		return err
	}

	body := &plan.Changes{}
	if err := c.BindBody(body); err != nil {
		return err
	}

	err := h.Provider.ApplyChanges(c.Request().Context(), body)
	if err != nil {
		return c.NewHTTPError(http.StatusUnprocessableEntity, err)
	}

	return c.NoContent(http.StatusOK)
}
