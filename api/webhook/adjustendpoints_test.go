package webhook_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/browningluke/opnsense-go/pkg/unbound"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/webhook"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/external-dns/endpoint"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("adjustendpoints", func() {
	Context("POST", func() {
		It("should be able to validate the incoming headers", func() {
			c, res := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodPost, "/", nil))

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).To(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusNotAcceptable))
		})

		It("should be able to handle when there is no item", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{})),
			)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("should be able to handle when there is no actionable item", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					endpoint.NewEndpoint("example.com", endpoint.RecordTypeA, "192.168.1.1"),
					endpoint.NewEndpoint("example.com", endpoint.RecordTypeAAAA, "fd00::"),
					endpoint.NewEndpoint("example.com", endpoint.RecordTypePTR),
					endpoint.NewEndpoint("example.com", endpoint.RecordTypeTXT),
				})),
			)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("should be able to handle when there is drifted actionable item", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					endpoint.NewEndpoint("example.com", endpoint.RecordTypeA, "192.168.1.1"),
					endpoint.NewEndpoint("example.com", endpoint.RecordTypeAAAA, "fd00::"),
					endpoint.NewEndpoint("example.com", endpoint.RecordTypePTR),
					endpoint.NewEndpoint("drifted.example.com", endpoint.RecordTypeA, "192.168.1.1").
						WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id").
						WithProviderSpecific(provider.ProviderSpecificDrifted.String(), "true"),
				})),
			)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundUpdateHostOverride(
				mock.Anything,
				"id",
				&unbound.HostOverride{
					Enabled:  "1",
					Hostname: "drifted",
					Domain:   "example.com",
					Type:     "A",
					Server:   "192.168.1.1",
				},
			).Return(nil)

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
		})
	})
})
