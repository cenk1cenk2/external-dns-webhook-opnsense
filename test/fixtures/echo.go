package fixtures

import (
	"net/http"
	"net/http/httptest"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/gomega"
)

func CreateEcho() *echo.Echo {
	e := echo.New()
	e.Validator = services.NewValidator()

	return e
}

func CreateEchoContext(e *echo.Echo, req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	if e == nil {
		e = CreateEcho()
	}

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c, rec
}

func GetEchoRouterContext(e *echo.Echo, req *http.Request, middlewares ...echo.MiddlewareFunc) (echo.Context, *httptest.ResponseRecorder, ctx.HandlerFunc) {
	c, rec := CreateEchoContext(e, req)

	e.Router().Find(req.Method, req.URL.Path, c)

	handler := c.Handler()
	Expect(handler).ToNot(BeNil())

	handler = ApplyMiddlewares(handler, middlewares)

	return c, rec, func(cc *ctx.Context) error {
		return handler(cc)
	}
}

func ApplyMiddlewares(handler echo.HandlerFunc, middlewares []echo.MiddlewareFunc) echo.HandlerFunc {
	if len(middlewares) == 0 {
		return handler
	}

	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}

func CreateEchoRouteGroup(e *echo.Echo, path string, middlewares ...echo.MiddlewareFunc) *echo.Group {
	if e == nil {
		e = CreateEcho()
	}

	return e.Group(path, middlewares...)
}
