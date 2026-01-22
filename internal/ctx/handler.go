package ctx

import (
	"github.com/labstack/echo/v5"
)

type HandlerFunc func(c *Context) error

func With(h HandlerFunc, with ...any) echo.HandlerFunc {
	return func(c *echo.Context) error {
		cc := NewContext(c, with...)

		return h(cc)
	}
}
