package api_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/metrics"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

var _ = Describe("Middleware", func() {
	var (
		a *api.Api
		m *metrics.Metrics
	)

	BeforeEach(func() {
		m = metrics.New()
		c := fixtures.NewTestConfig()
		logger := fixtures.NewTestLogger()
		validator := services.NewValidator()

		a = api.NewApi(&api.ApiSvc{
			Logger:    logger,
			Validator: validator,
			Metrics:   m,
		}, c.Api)
		Expect(a).ToNot(BeNil())
		Expect(a.Echo).ToNot(BeNil())
	})

	It("should be able to recover from a panic", func() {
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)

		a.Echo.GET("/panic", ctx.With(
			func(c *ctx.Context) error {
				panic("imdat")
			}),
		)

		c, res, h := fixtures.GetEchoRouterContext(a.Echo, req, a.GetMiddlewares()...)

		// In Echo v5, panic recovery returns an error
		err := fixtures.Respond(c, h)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("PANIC RECOVER"))
		Expect(res.Code).To(Equal(http.StatusInternalServerError))
	})

	Context("metrics", func() {
		It("should record requests by route and status", func() {
			a.Echo.GET("/records", ctx.With(
				func(c *ctx.Context) error {
					return c.NoContent(http.StatusAccepted)
				}),
			)

			c, _, h := fixtures.GetEchoRouterContext(a.Echo, httptest.NewRequest(http.MethodGet, "/records", nil), a.GetMiddlewares()...)
			Expect(fixtures.Respond(c, h)).To(Succeed())

			Expect(testutil.ToFloat64(m.HttpRequests.WithLabelValues(http.MethodGet, "/records", "202"))).To(Equal(1.0))
			Expect(testutil.CollectAndCount(m.HttpRequestDuration)).To(Equal(1))
		})

		It("should record the status of a failed request", func() {
			a.Echo.GET("/fail", ctx.With(
				func(c *ctx.Context) error {
					return echo.NewHTTPError(http.StatusBadRequest, "nope")
				}),
			)

			c, _, h := fixtures.GetEchoRouterContext(a.Echo, httptest.NewRequest(http.MethodGet, "/fail", nil), a.GetMiddlewares()...)
			Expect(fixtures.Respond(c, h)).To(HaveOccurred())

			Expect(testutil.ToFloat64(m.HttpRequests.WithLabelValues(http.MethodGet, "/fail", "400"))).To(Equal(1.0))
		})

		It("should not record probe paths", func() {
			a.Echo.GET("/healthz", ctx.With(
				func(c *ctx.Context) error {
					return c.NoContent(http.StatusOK)
				}),
			)

			c, _, h := fixtures.GetEchoRouterContext(a.Echo, httptest.NewRequest(http.MethodGet, "/healthz", nil), a.GetMiddlewares()...)
			Expect(fixtures.Respond(c, h)).To(Succeed())

			Expect(testutil.CollectAndCount(m.HttpRequests)).To(Equal(0))
		})
	})
})
