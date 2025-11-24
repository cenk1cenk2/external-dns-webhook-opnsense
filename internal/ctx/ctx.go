package ctx

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/interfaces"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/labstack/echo/v4"
)

type Context struct {
	echo.Context
	Log    services.ZapSugaredLogger
	binder *echo.DefaultBinder
}

type BindTarget struct {
	Path    interface{}
	Query   interface{}
	Headers interface{}
	Body    interface{}
}

func NewContext(c echo.Context, with ...interface{}) *Context {
	cc := &Context{
		Context: c,
		binder:  &echo.DefaultBinder{},
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

func (c *Context) BindPathParams(i interface{}) error {
	if err := c.binder.BindPathParams(c.Context, i); err != nil {
		return fmt.Errorf("failed to bind path parameters: %w", err)
	}

	if err := c.Validate(i); err != nil {
		return c.NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

func (c *Context) BindQueryParams(i interface{}) error {
	if err := c.binder.BindQueryParams(c.Context, i); err != nil {
		return fmt.Errorf("failed to bind query parameters: %w", err)
	}

	if err := c.Validate(i); err != nil {
		return c.NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

func (c *Context) BindHeaders(i interface{}) error {
	if err := c.binder.BindHeaders(c.Context, i); err != nil {
		return fmt.Errorf("failed to bind headers: %w", err)
	}

	if err := c.Validate(i); err != nil {
		return c.NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

func (c *Context) BindBody(i interface{}) error {
	if err := c.binder.BindBody(c.Context, i); err != nil {
		return fmt.Errorf("failed to bind body: %w", err)
	}

	if err := c.Validate(i); err != nil {
		return c.NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

func (c *Context) With(with ...interface{}) *Context {
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
