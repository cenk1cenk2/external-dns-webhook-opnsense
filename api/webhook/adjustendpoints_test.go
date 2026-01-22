package webhook_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/webhook"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	"github.com/labstack/echo/v5"
	"sigs.k8s.io/external-dns/endpoint"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("adjustendpoints", func() {
	Context("POST", func() {
		It("should be able to validate the incoming headers", func() {
			c, res := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodPost, "/", nil))

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).To(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusUnsupportedMediaType))
		})

		It("should be able to handle when there is no item", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Header().Get(echo.HeaderContentType)).To(Equal(webhook.ExternalDnsAcceptedMedia))
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
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Header().Get(echo.HeaderContentType)).To(Equal(webhook.ExternalDnsAcceptedMedia))
		})

		It("should add SetIdentifier to single target A record", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "app.example.com",
						RecordType: endpoint.RecordTypeA,
						Targets:    []string{"10.0.0.1"},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			body := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(body).To(HaveLen(1))
			Expect(body[0].DNSName).To(Equal("app.example.com"))
			Expect(body[0].Targets).To(ConsistOf("10.0.0.1"))
			Expect(body[0].SetIdentifier).ToNot(BeEmpty())
		})

		It("should split multiple targets into separate endpoints with unique SetIdentifiers", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "app.example.com",
						RecordType: endpoint.RecordTypeA,
						Targets:    []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			body := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(body).To(HaveLen(3))

			// All should have same DNSName and RecordType
			for _, ep := range body {
				Expect(ep.DNSName).To(Equal("app.example.com"))
				Expect(ep.RecordType).To(Equal(endpoint.RecordTypeA))
				Expect(ep.Targets).To(HaveLen(1))
				Expect(ep.SetIdentifier).ToNot(BeEmpty())
			}

			// Each should have unique SetIdentifier
			setIdentifiers := make(map[string]bool)
			for _, ep := range body {
				setIdentifiers[ep.SetIdentifier] = true
			}
			Expect(setIdentifiers).To(HaveLen(3))

			// Verify all targets present
			targets := []string{body[0].Targets[0], body[1].Targets[0], body[2].Targets[0]}
			Expect(targets).To(ConsistOf("10.0.0.1", "10.0.0.2", "10.0.0.3"))
		})

		It("should split TXT records with multiple values", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "example.com",
						RecordType: endpoint.RecordTypeTXT,
						Targets:    []string{"v=spf1 include:_spf.example.com ~all", "google-site-verification=abc123"},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			body := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(body).To(HaveLen(2))

			for _, ep := range body {
				Expect(ep.DNSName).To(Equal("example.com"))
				Expect(ep.RecordType).To(Equal(endpoint.RecordTypeTXT))
				Expect(ep.Targets).To(HaveLen(1))
				Expect(ep.SetIdentifier).ToNot(BeEmpty())
			}

			txtValues := []string{body[0].Targets[0], body[1].Targets[0]}
			Expect(txtValues).To(ConsistOf("v=spf1 include:_spf.example.com ~all", "google-site-verification=abc123"))
		})

		It("should preserve ProviderSpecific properties when splitting", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "app.example.com",
						RecordType: endpoint.RecordTypeA,
						Targets:    []string{"10.0.0.1", "10.0.0.2"},
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "custom-key", Value: "custom-value"},
						},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			body := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(body).To(HaveLen(2))

			for _, ep := range body {
				Expect(ep.ProviderSpecific).To(HaveLen(1))
				value, exists := ep.GetProviderSpecificProperty("custom-key")
				Expect(exists).To(BeTrue())
				Expect(value).To(Equal("custom-value"))
			}
		})

		It("should preserve Labels and RecordTTL when splitting", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "app.example.com",
						RecordType: endpoint.RecordTypeA,
						Targets:    []string{"10.0.0.1", "10.0.0.2"},
						RecordTTL:  endpoint.TTL(300),
						Labels: map[string]string{
							"owner": "team-a",
						},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			body := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(body).To(HaveLen(2))

			for _, ep := range body {
				Expect(ep.RecordTTL).To(Equal(endpoint.TTL(300)))
				Expect(ep.Labels).To(HaveKeyWithValue("owner", "team-a"))
			}
		})

		It("should handle AAAA records", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "app.example.com",
						RecordType: endpoint.RecordTypeAAAA,
						Targets:    []string{"fd00::1", "fd00::2"},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			body := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(body).To(HaveLen(2))

			for _, ep := range body {
				Expect(ep.RecordType).To(Equal(endpoint.RecordTypeAAAA))
				Expect(ep.SetIdentifier).ToNot(BeEmpty())
			}

			targets := []string{body[0].Targets[0], body[1].Targets[0]}
			Expect(targets).To(ConsistOf("fd00::1", "fd00::2"))
		})

		It("should handle mixed endpoints - some with SetIdentifier, some without", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:       "existing.example.com",
						RecordType:    endpoint.RecordTypeA,
						Targets:       []string{"10.0.0.1"},
						SetIdentifier: "existing-id",
					},
					{
						DNSName:    "new.example.com",
						RecordType: endpoint.RecordTypeA,
						Targets:    []string{"10.0.0.2", "10.0.0.3"},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(fixtures.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			body := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(body).To(HaveLen(3))

			// First endpoint has regenerated SetIdentifier (not preserved)
			Expect(body[0].DNSName).To(Equal("existing.example.com"))
			Expect(body[0].SetIdentifier).ToNot(BeEmpty())

			// Next two are split from second input
			Expect(body[1].DNSName).To(Equal("new.example.com"))
			Expect(body[2].DNSName).To(Equal("new.example.com"))
			Expect(body[1].SetIdentifier).ToNot(BeEmpty())
			Expect(body[2].SetIdentifier).ToNot(BeEmpty())
			Expect(body[1].SetIdentifier).ToNot(Equal(body[2].SetIdentifier))
		})

		It("should generate stable SetIdentifiers - same input produces same output", func() {
			inputEndpoints := []*endpoint.Endpoint{
				{
					DNSName:    "stable.example.com",
					RecordType: endpoint.RecordTypeA,
					Targets:    []string{"10.0.0.1"},
				},
			}

			// First call
			req1 := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal(inputEndpoints)),
			)
			req1.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c1, res1 := fixtures.CreateEchoContext(nil, req1)

			Expect(fixtures.Respond(c1, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			body1 := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res1.Body.Bytes())

			// Second call with same input
			req2 := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal(inputEndpoints)),
			)
			req2.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c2, res2 := fixtures.CreateEchoContext(nil, req2)

			Expect(fixtures.Respond(c2, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			body2 := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res2.Body.Bytes())

			// SetIdentifiers should be identical
			Expect(body1[0].SetIdentifier).To(Equal(body2[0].SetIdentifier))
		})
	})
})
