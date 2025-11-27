package probes

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/interfaces"
	"github.com/karagenc/zap4echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap/zapcore"
)

func (a *Api) SetupMiddleware() {
	e := a.Echo
	e.Validator = a.Validator

	e.OnAddRouteHandler = func(host string, route echo.Route, handler echo.HandlerFunc, middleware []echo.MiddlewareFunc) {
		a.log.Debugf("Registered route: %s %s -> %s", route.Method, route.Path, route.Name)
	}

	e.Use(a.GetMiddlewares()...)

	if a.Logger.Level() <= zapcore.DebugLevel {
		a.log.Warnln("Enabled debug mode in echo.")
		e.Debug = true
	}

	e.HTTPErrorHandler = a.HTTPErrorHandler
}

func (a *Api) GetMiddlewares() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		zap4echo.RecoverWithConfig(a.Logger.Logger, zap4echo.RecoverConfig{
			CustomRequestIDHeader: echo.HeaderXRequestID,
		}),
		zap4echo.LoggerWithConfig(a.Logger.Logger, zap4echo.LoggerConfig{
			CustomRequestIDHeader: echo.HeaderXRequestID,
			IncludeCaller:         true,
			OmitReferer:           true,
			Skipper: func(c echo.Context) bool {
				return slices.Contains([]bool{
					strings.HasPrefix(c.Path(), "/healthz"),
					strings.HasPrefix(c.Path(), "/readyz"),
				}, true)
			},
		}),
		middleware.RequestID(),
		middleware.CORS(),
	}
}

func (a *Api) HTTPErrorHandler(err error, c echo.Context) {
	var e *echo.HTTPError
	if ok := errors.As(err, &e); ok {
		if cast, o := e.Message.(error); o {
			_ = c.JSON(e.Code, interfaces.ApiError{
				Status:  e.Code,
				Message: cast.Error(),
			})
		} else {
			_ = c.JSON(e.Code, interfaces.ApiError{
				Status:  e.Code,
				Message: e.Message.(string),
			})
		}

		return
	}

	_ = c.JSON(http.StatusInternalServerError, interfaces.ApiError{
		Status:  http.StatusInternalServerError,
		Message: err.Error(),
	})
}
