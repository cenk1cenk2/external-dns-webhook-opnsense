package webhook_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/webhook"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	"github.com/labstack/echo/v5"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("negotiate", func() {
	Context("GET", func() {
		It("should be able to validate the incoming headers", func() {
			c, res := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodGet, "/", nil))

			Expect(ctx.Respond(c, handler.HandleNegotiateGet)).To(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusNotAcceptable))
		})

		It("should be able the negotiate with incoming request", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleNegotiateGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Header().Get(echo.HeaderContentType)).To(Equal(webhook.ExternalDnsAcceptedMedia))
			Expect(res.Body).To(MatchJSON(`{}`))
		})

		It("should be able to negotiate with direct config", func() {
			handler.Provider.DomainFilter = provider.NewDomainFilter(provider.DomainFilterConfig{
				DomainFilter:   []string{"example.com"},
				ExcludeDomains: []string{"excluded.example.com"},
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleNegotiateGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Header().Get(echo.HeaderContentType)).To(Equal(webhook.ExternalDnsAcceptedMedia))
			Expect(res.Body).To(MatchJSON(`{
        "include": [
          "example.com"
        ],
        "exclude": [
          "excluded.example.com"
        ]
      }`))
		})

		It("should be able to negotiate with regex config", func() {
			handler.Provider.DomainFilter = provider.NewDomainFilter(provider.DomainFilterConfig{
				RegexDomainFilter:    `^.*\.example\.com$`,
				RegexDomainExclusion: `^excluded.example.com$`,
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleNegotiateGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Header().Get(echo.HeaderContentType)).To(Equal(webhook.ExternalDnsAcceptedMedia))
			Expect(res.Body).To(MatchJSON(`{
        "regexInclude": "^.*\\.example\\.com$",
        "regexExclude": "^excluded.example.com$"
      }`))
		})
	})
})
