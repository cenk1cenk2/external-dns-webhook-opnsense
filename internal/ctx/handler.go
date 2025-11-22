package ctx

import (
	"github.com/labstack/echo/v4"
)

type HandlerFunc func(c *Context) error

func With(h HandlerFunc, with ...interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		return Run(c, h, with...)
	}
}

func Run(c echo.Context, h HandlerFunc, with ...interface{}) error {
	cc := NewContext(c, with...)

	return h(cc)
}

func Respond(c echo.Context, h HandlerFunc, with ...interface{}) error {
	err := Run(c, h, with...)
	if err != nil {
		c.Echo()
		c.Error(err)
	}
	c.Response().Flush()

	return err
}
