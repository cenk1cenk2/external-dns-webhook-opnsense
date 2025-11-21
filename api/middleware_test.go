package api_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Middleware", func() {
	var a *api.Api

	BeforeEach(func() {
		c := fixtures.NewTestConfig()
		logger := fixtures.NewTestLogger()
		validator := services.NewValidator()

		a = api.NewApi(&api.ApiSvc{
			Log:       logger,
			Validator: validator,
		}, c.Api)
		Expect(a).ToNot(BeNil())
		Expect(a.Echo).ToNot(BeNil())
	})

	It("should be able to recover from a panic", func() {
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)

		a.Echo.GET("/panic", ctx.WrapHandler(
			func(c *ctx.Context) error {
				panic("imdat")
			}),
		)

		c, res, h := fixtures.GetEchoRouterContext(a.Echo, req, a.GetMiddlewares()...)

		Expect(ctx.RespondWithContext(c, h)).ToNot(HaveOccurred())
		Expect(res.Code).To(Equal(http.StatusInternalServerError))
	})
})
