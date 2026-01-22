package ctx

import (
	"github.com/labstack/echo/v5"
)

type HandlerFunc func(c *Context) error

func With(h HandlerFunc, with ...any) echo.HandlerFunc {
	return func(c *echo.Context) error {
		return Run(c, h, with...)
	}
}

func Run(c *echo.Context, h HandlerFunc, with ...any) error {
	cc := NewContext(c, with...)

	return h(cc)
}

func Respond(c *echo.Context, h HandlerFunc, with ...any) error {
	err := Run(c, h, with...)

	if err != nil {
		c.Echo().HTTPErrorHandler(c, err)
	}

	return err
}
