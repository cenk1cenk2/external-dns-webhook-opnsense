package fixtures

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/labstack/echo/v5"
)

func Run(c *echo.Context, h ctx.HandlerFunc, with ...any) error {
	cc := ctx.NewContext(c, with...)

	return h(cc)
}

func Respond(c *echo.Context, h ctx.HandlerFunc, with ...any) error {
	err := Run(c, h, with...)

	if err != nil {
		c.Echo().HTTPErrorHandler(c, err)
	}

	return err
}
