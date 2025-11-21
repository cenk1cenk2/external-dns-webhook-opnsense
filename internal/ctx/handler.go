package ctx

import (
	"github.com/labstack/echo/v4"
)

type HandlerFunc func(c *Context) error

func WrapHandler(h HandlerFunc, with ...interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		return RunWithContext(c, h, with...)
	}
}

func RunWithContext(c echo.Context, h HandlerFunc, with ...interface{}) error {
	cc := NewContext(c, with...)

	return h(cc)
}

func RespondWithContext(c echo.Context, h HandlerFunc, with ...interface{}) error {
	err := RunWithContext(c, h, with...)
	if err != nil {
		c.Echo()
		c.Error(err)
	}
	c.Response().Flush()

	return err
}
