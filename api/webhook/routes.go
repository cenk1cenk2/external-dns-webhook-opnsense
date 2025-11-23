package webhook

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) RegisterRoutes(r *echo.Group) *Handler {
	// g := r.Group("")

	// TODO: GET / negotiate
	// TODO: GET /records read records
	// TODO: POST /records set records
	// TODO: POST /adjustendpoints adjust endpoints

	return h
}
