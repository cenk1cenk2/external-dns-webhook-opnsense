package ctx

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/interfaces"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/labstack/echo/v5"
)

type Context struct {
	*echo.Context
	Log services.ZapSugaredLogger
}

type BindTarget struct {
	Path    any
	Query   any
	Headers any
	Body    any
}

func NewContext(c *echo.Context, with ...any) *Context {
	cc := &Context{
		Context: c,
	}

	if len(with) > 0 {
		cc.With(with...)
	}

	return cc
}

func (c *Context) BindTarget(target *BindTarget) error {
	if target.Path != nil {
		if err := c.BindPathParams(target.Path); err != nil {
			return err
		}
	}

	if target.Query != nil {
		if err := c.BindQueryParams(target.Query); err != nil {
			return err
		}
	}

	if target.Headers != nil {
		if err := c.BindHeaders(target.Headers); err != nil {
			return err
		}
	}

	if target.Body != nil {
		if err := c.BindBody(target.Body); err != nil {
			return err
		}
	}

	return nil
}

func (c *Context) BindPathParams(i any) error {
	if err := echo.BindPathValues(c.Context, i); err != nil {
		return fmt.Errorf("failed to bind path parameters: %w", err)
	}

	if err := c.Validate(i); err != nil {
		return c.NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

func (c *Context) BindQueryParams(i any) error {
	if err := echo.BindQueryParams(c.Context, i); err != nil {
		return fmt.Errorf("failed to bind query parameters: %w", err)
	}

	if err := c.Validate(i); err != nil {
		return c.NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

func (c *Context) BindHeaders(i any) error {
	if err := echo.BindHeaders(c.Context, i); err != nil {
		return fmt.Errorf("failed to bind headers: %w", err)
	}

	if err := c.Validate(i); err != nil {
		return c.NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

func (c *Context) BindBody(i any) error {
	if err := echo.BindBody(c.Context, i); err != nil {
		return fmt.Errorf("failed to bind body: %w", err)
	}

	if err := c.Validate(i); err != nil {
		return c.NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

func (c *Context) With(with ...any) *Context {
	for _, v := range with {
		switch w := v.(type) {
		case *services.Logger:
			c.Log = w.WithEchoContext(c.Context)
		default:
			panic(fmt.Sprintf("Can not handle with: %s", reflect.TypeOf(w)))
		}
	}

	return c
}

func (c *Context) NewHTTPError(code int, err error) error {
	return interfaces.NewHttpError(code, err)
}
