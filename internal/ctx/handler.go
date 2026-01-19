package ctx

import (
	"github.com/labstack/echo/v5"
)

type HandlerFunc func(c *Context) error

func With(h HandlerFunc, with ...interface{}) echo.HandlerFunc {
	return func(c *echo.Context) error {
		return Run(c, h, with...)
	}
}

func Run(c *echo.Context, h HandlerFunc, with ...interface{}) error {
	cc := NewContext(c, with...)

	return h(cc)
}

func Respond(c *echo.Context, h HandlerFunc, with ...interface{}) error {
	err := Run(c, h, with...)

	// Check if response was already committed (e.g., by panic recovery middleware)
	if res, unwrapped := echo.UnwrapResponse(c.Response()); unwrapped == nil && res.Committed {
		// Response already sent, don't call error handler
		return err
	}

	if err != nil {
		c.Echo().HTTPErrorHandler(c, err)
	}

	if flusher, ok := c.Response().(interface{ Flush() }); ok {
		flusher.Flush()
	}

	return err
}
