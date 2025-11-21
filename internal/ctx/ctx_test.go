package ctx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	"github.com/labstack/echo/v4"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context", func() {
	var e *echo.Echo

	BeforeEach(func() {
		e = fixtures.CreateEcho()
	})

	It("should be able to bind path parameters", func() {
		type PathParams struct {
			A string `param:"a"`
			B string `param:"b"`
		}

		req := httptest.NewRequest(
			http.MethodGet,
			"/test/path",
			nil,
		)

		e.GET("/:a/:b", ctx.WrapHandler(
			func(c *ctx.Context) error {
				params := &PathParams{}

				err := c.BindPathParams(params)

				Expect(err).ToNot(HaveOccurred())
				Expect(params.A).To(Equal("test"))
				Expect(params.B).To(Equal("path"))

				return nil
			}),
		)

		c, res, h := fixtures.GetEchoRouterContext(e, req)

		Expect(ctx.RespondWithContext(c, h)).ToNot(HaveOccurred())
		Expect(res.Code).To(Equal(http.StatusOK))
	})

	It("should be able to bind query parameters", func() {
		type QueryParams struct {
			A string `query:"a"`
			B string `query:"b"`
		}

		req := httptest.NewRequest(
			http.MethodGet,
			"/test/path?a=test&b=path",
			nil,
		)
		q := req.URL.Query()
		q.Add("a", "test")
		q.Add("b", "path")
		req.URL.RawQuery = q.Encode()

		e.GET("/:a/:b", ctx.WrapHandler(
			func(c *ctx.Context) error {
				params := &QueryParams{}

				err := c.BindQueryParams(params)

				Expect(err).ToNot(HaveOccurred())
				Expect(params.A).To(Equal("test"))
				Expect(params.B).To(Equal("path"))

				return nil
			}),
		)

		c, res, h := fixtures.GetEchoRouterContext(e, req)

		Expect(ctx.RespondWithContext(c, h)).ToNot(HaveOccurred())
		Expect(res.Code).To(Equal(http.StatusOK))
	})

	It("should be able to bind headers", func() {
		type Headers struct {
			A string `header:"a"`
			B string `header:"b"`
		}

		req := httptest.NewRequest(
			http.MethodGet,
			"/test/path",
			nil,
		)
		req.Header.Add("a", "test")
		req.Header.Add("b", "path")

		e.GET("/:a/:b", ctx.WrapHandler(
			func(c *ctx.Context) error {
				params := &Headers{}

				err := c.BindHeaders(params)

				Expect(err).ToNot(HaveOccurred())
				Expect(params.A).To(Equal("test"))
				Expect(params.B).To(Equal("path"))

				return nil
			}),
		)

		c, res, h := fixtures.GetEchoRouterContext(e, req)

		Expect(ctx.RespondWithContext(c, h)).ToNot(HaveOccurred())
		Expect(res.Code).To(Equal(http.StatusOK))
	})

	It("should be able to bind body", func() {
		type Body struct {
			Name string `json:"name"`
		}

		req := fixtures.SetRequestContentJson(
			httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(
					fixtures.MustJsonMarshal(&Body{
						Name: "test",
					}),
				),
			),
		)

		e.POST("/", ctx.WrapHandler(
			func(c *ctx.Context) error {
				body := &Body{}

				err := c.BindBody(body)

				Expect(err).ToNot(HaveOccurred())
				Expect(body.Name).To(Equal("test"))

				return nil
			}),
		)

		c, res, h := fixtures.GetEchoRouterContext(e, req)

		Expect(ctx.RespondWithContext(c, h)).ToNot(HaveOccurred())
		Expect(res.Code).To(Equal(http.StatusOK))
	})

	It("should be able to bind target", func() {
		type PathParams struct {
			A string `param:"a"`
			B string `param:"b"`
		}

		type QueryParams struct {
			A string `query:"a"`
		}

		type Headers struct {
			A string `header:"a"`
		}

		type Body struct {
			Name string `json:"name"`
		}

		type Req struct {
			Path    *PathParams
			Query   *QueryParams
			Headers *Headers
			Body    *Body
		}

		req := fixtures.SetRequestContentJson(
			httptest.NewRequest(
				http.MethodPost,
				"/test/path",
				strings.NewReader(
					fixtures.MustJsonMarshal(&Body{
						Name: "test",
					}),
				),
			),
		)
		q := req.URL.Query()
		q.Add("a", "test")
		req.URL.RawQuery = q.Encode()
		req.Header.Add("a", "test")

		e.POST("/:a/:b", ctx.WrapHandler(
			func(c *ctx.Context) error {
				req := &Req{
					Path:    &PathParams{},
					Query:   &QueryParams{},
					Headers: &Headers{},
					Body:    &Body{},
				}

				err := c.BindTarget(&ctx.BindTarget{
					Path:    req.Path,
					Query:   req.Query,
					Headers: req.Headers,
					Body:    req.Body,
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(req.Path.A).To(Equal("test"))
				Expect(req.Path.B).To(Equal("path"))
				Expect(req.Query.A).To(Equal("test"))
				Expect(req.Headers.A).To(Equal("test"))
				Expect(req.Body.Name).To(Equal("test"))

				return nil
			}),
		)

		c, res, h := fixtures.GetEchoRouterContext(e, req)

		Expect(ctx.RespondWithContext(c, h)).ToNot(HaveOccurred())
		Expect(res.Code).To(Equal(http.StatusOK))
	})

	It("should be able to wrap a context with logger", func() {
		req := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

		e.GET("/", ctx.WrapHandler(
			func(c *ctx.Context) error {
				c.Log.Warnf("yattara")

				return nil
			},
			fixtures.NewTestLogger(),
		))

		c, res, h := fixtures.GetEchoRouterContext(e, req)

		Expect(ctx.RespondWithContext(c, h)).ToNot(HaveOccurred())
		Expect(res.Code).To(Equal(http.StatusOK))
	})
})
