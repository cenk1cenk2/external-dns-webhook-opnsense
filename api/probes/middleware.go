package probes

import (
	"errors"
	"net/http"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/interfaces"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func (a *Api) SetupMiddleware() {
	e := a.Echo
	e.Validator = a.Validator

	e.OnAddRoute = func(route echo.Route) error {
		a.log.Debugf("Registered route: %s %s -> %s", route.Method, route.Path, route.Name)

		return nil
	}

	e.Use(a.GetMiddlewares()...)

	e.HTTPErrorHandler = a.HTTPErrorHandler
}

func (a *Api) GetMiddlewares() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		middleware.Recover(),
		middleware.RequestID(),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
		}),
	}
}

func (a *Api) HTTPErrorHandler(c *echo.Context, err error) {
	var e *echo.HTTPError
	if ok := errors.As(err, &e); ok {
		_ = c.JSON(e.Code, interfaces.ApiError{
			Status:  e.Code,
			Message: e.Message,
		})

		return
	}

	_ = c.JSON(http.StatusInternalServerError, interfaces.ApiError{
		Status:  http.StatusInternalServerError,
		Message: err.Error(),
	})
}
