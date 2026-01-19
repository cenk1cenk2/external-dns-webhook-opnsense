package api

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/interfaces"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func (a *Api) SetupMiddleware() {
	e := a.Echo
	e.Validator = a.Validator

	e.OnAddRoute = func(route echo.Route) error {
		a.log.Debugf("Registered route: %s %s", route.Method, route.Path)

		return nil
	}

	e.Use(a.GetMiddlewares()...)

	e.HTTPErrorHandler = a.HTTPErrorHandler
}

func (a *Api) GetMiddlewares() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		middleware.Recover(),
		middleware.RequestID(),
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogStatus:   true,
			LogURI:      true,
			LogMethod:   true,
			LogLatency:  true,
			LogRemoteIP: true,
			Skipper: func(c *echo.Context) bool {
				return slices.Contains([]bool{
					strings.HasPrefix(c.Path(), "/healthz"),
					strings.HasPrefix(c.Path(), "/readyz"),
				}, true)
			},
			LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
				logger := a.Logger.WithEchoContext(c)
				if v.Error != nil {
					logger.Error(v.Error.Error())
				} else {
					logger.Info("request")
				}

				return nil
			},
		}),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
		}),
	}
}

func (a *Api) HTTPErrorHandler(c *echo.Context, err error) {
	log := a.Logger.WithEchoContext(c)

	var e *echo.HTTPError
	if ok := errors.As(err, &e); ok {
		log.Errorf("HTTP %d - %s", e.Code, e.Message)
		_ = c.JSON(e.Code, interfaces.ApiError{
			Status:  e.Code,
			Message: e.Message,
		})

		return
	}

	log.Errorf("HTTP %d - %s", http.StatusInternalServerError, err.Error())
	_ = c.JSON(http.StatusInternalServerError, interfaces.ApiError{
		Status:  http.StatusInternalServerError,
		Message: err.Error(),
	})
}
