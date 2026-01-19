package webhook_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/browningluke/opnsense-go/pkg/unbound"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/webhook"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

var _ = Describe("records", func() {
	Context("GET", func() {
		It("should be able to validate the incoming headers", func() {
			c, res := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodGet, "/", nil))

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).To(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusNotAcceptable))
		})

		It("should be able to handle errors while fetching the records", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(nil, fmt.Errorf("")).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).To(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusUnprocessableEntity))
		})

		It("should be able to fetch the records on empty response", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(&unbound.SearchHostOverrideResponse{}, nil).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Header().Get(echo.HeaderContentType)).To(Equal(webhook.ExternalDnsAcceptedMedia))

			var body []endpoint.Endpoint
			Expect(json.Unmarshal(res.Body.Bytes(), &body)).To(Succeed())
			Expect(body).To(BeEmpty())
		})

		It("should be able to fetch and convert all the records", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{
					Total:    1,
					RowCount: 1,
					Current:  1,
					Rows: []unbound.SearchHostOverrideItem{
						{
							Id:       "id",
							Enabled:  "1",
							Hostname: "example",
							Domain:   "com",
							Type:     "A",
							Server:   "192.168.1.1",
						},
					},
				},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Header().Get(echo.HeaderContentType)).To(Equal(webhook.ExternalDnsAcceptedMedia))

			var body []endpoint.Endpoint
			Expect(json.Unmarshal(res.Body.Bytes(), &body)).To(Succeed())
			Expect(body).To(HaveLen(1))

			Expect(body[0].DNSName).To(Equal("example.com"))
			Expect(body[0].Targets).To(BeEquivalentTo([]string{"192.168.1.1"}))
			Expect(body[0].RecordType).To(Equal("A"))
			Expect(body[0].SetIdentifier).ToNot(BeEmpty()) // Stable hash based on record data
			Expect(body[0].ProviderSpecific).To(ContainElements(
				endpoint.ProviderSpecificProperty{
					Name:  provider.ProviderSpecificUUID.String(),
					Value: "id",
				},
			))
		})

		It("should be able to fetch and convert TXT records", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{
					Total:    1,
					RowCount: 1,
					Current:  1,
					Rows: []unbound.SearchHostOverrideItem{
						{
							Id:      "id-txt",
							Enabled: "1",
							Type:    "TXT",
							Domain:  "a-example.com",
							TxtData: "heritage=external-dns,external-dns/owner=test-cluster",
						},
					},
				},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Header().Get(echo.HeaderContentType)).To(Equal(webhook.ExternalDnsAcceptedMedia))

			var body []endpoint.Endpoint
			Expect(json.Unmarshal(res.Body.Bytes(), &body)).To(Succeed())
			Expect(body).To(HaveLen(1))

			Expect(body[0].DNSName).To(Equal("a-example.com"))
			Expect(body[0].Targets).To(BeEquivalentTo([]string{"heritage=external-dns,external-dns/owner=test-cluster"}))
			Expect(body[0].RecordType).To(Equal("TXT"))
			Expect(body[0].ProviderSpecific).To(ContainElements(
				endpoint.ProviderSpecificProperty{
					Name:  provider.ProviderSpecificUUID.String(),
					Value: "id-txt",
				},
			))
		})

		It("should be able to fetch records with descriptions", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{
					Total:    1,
					RowCount: 1,
					Current:  1,
					Rows: []unbound.SearchHostOverrideItem{
						{
							Id:          "id-with-desc",
							Enabled:     "1",
							Hostname:    "api",
							Domain:      "example.com",
							Type:        "A",
							Server:      "192.168.1.50",
							Description: "Production API endpoint",
						},
					},
				},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			var body []endpoint.Endpoint
			Expect(json.Unmarshal(res.Body.Bytes(), &body)).To(Succeed())
			Expect(body).To(HaveLen(1))

			Expect(body[0].DNSName).To(Equal("api.example.com"))
			Expect(body[0].SetIdentifier).ToNot(BeEmpty()) // Stable hash based on record data
			Expect(body[0].ProviderSpecific).To(ContainElements(
				endpoint.ProviderSpecificProperty{
					Name:  provider.ProviderSpecificUUID.String(),
					Value: "id-with-desc",
				},
				endpoint.ProviderSpecificProperty{
					Name:  provider.ProviderSpecificDescription.String(),
					Value: "Production API endpoint",
				},
			))
		})

		It("should be able to fetch mixed A and TXT records", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{
					Total:    2,
					RowCount: 2,
					Current:  1,
					Rows: []unbound.SearchHostOverrideItem{
						{
							Id:       "id-a",
							Enabled:  "1",
							Type:     "A",
							Hostname: "example",
							Domain:   "com",
							Server:   "192.168.1.1",
						},
						{
							Id:      "id-txt",
							Enabled: "1",
							Type:    "TXT",
							Domain:  "a-example.com",
							TxtData: "heritage=external-dns,external-dns/owner=test-cluster",
						},
					},
				},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Header().Get(echo.HeaderContentType)).To(Equal(webhook.ExternalDnsAcceptedMedia))

			var body []endpoint.Endpoint
			Expect(json.Unmarshal(res.Body.Bytes(), &body)).To(Succeed())
			Expect(body).To(HaveLen(2))

			Expect(body[0].DNSName).To(Equal("example.com"))
			Expect(body[0].Targets).To(BeEquivalentTo([]string{"192.168.1.1"}))
			Expect(body[0].RecordType).To(Equal("A"))

			Expect(body[1].DNSName).To(Equal("a-example.com"))
			Expect(body[1].Targets).To(BeEquivalentTo([]string{"heritage=external-dns,external-dns/owner=test-cluster"}))
			Expect(body[1].RecordType).To(Equal("TXT"))
		})

		It("should be able to fetch multiple A records with same hostname as separate endpoints", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{
					Total:    3,
					RowCount: 3,
					Current:  1,
					Rows: []unbound.SearchHostOverrideItem{
						{
							Id:       "id-a-1",
							Enabled:  "1",
							Type:     "A",
							Hostname: "example",
							Domain:   "com",
							Server:   "192.168.1.1",
						},
						{
							Id:       "id-a-2",
							Enabled:  "1",
							Type:     "A",
							Hostname: "example",
							Domain:   "com",
							Server:   "192.168.1.2",
						},
						{
							Id:       "id-a-3",
							Enabled:  "1",
							Type:     "A",
							Hostname: "example",
							Domain:   "com",
							Server:   "192.168.1.3",
						},
					},
				},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			var body []endpoint.Endpoint
			Expect(json.Unmarshal(res.Body.Bytes(), &body)).To(Succeed())
			Expect(body).To(HaveLen(3))

			Expect(body[0].DNSName).To(Equal("example.com"))
			Expect(body[0].Targets).To(ConsistOf("192.168.1.1"))
			Expect(body[0].RecordType).To(Equal("A"))
			Expect(body[0].SetIdentifier).ToNot(BeEmpty()) // Stable hash: example.com:A:192.168.1.1

			Expect(body[1].DNSName).To(Equal("example.com"))
			Expect(body[1].Targets).To(ConsistOf("192.168.1.2"))
			Expect(body[1].RecordType).To(Equal("A"))
			Expect(body[1].SetIdentifier).ToNot(BeEmpty()) // Stable hash: example.com:A:192.168.1.2

			Expect(body[2].DNSName).To(Equal("example.com"))
			Expect(body[2].Targets).To(ConsistOf("192.168.1.3"))
			Expect(body[2].RecordType).To(Equal("A"))
			Expect(body[2].SetIdentifier).ToNot(BeEmpty()) // Stable hash: example.com:A:192.168.1.3

			// Each record should have different SetIdentifiers because targets differ
			Expect(body[0].SetIdentifier).ToNot(Equal(body[1].SetIdentifier))
			Expect(body[1].SetIdentifier).ToNot(Equal(body[2].SetIdentifier))
			Expect(body[0].SetIdentifier).ToNot(Equal(body[2].SetIdentifier))
		})

		It("should be able to fetch multiple TXT records with same domain as separate endpoints", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{
					Total:    2,
					RowCount: 2,
					Current:  1,
					Rows: []unbound.SearchHostOverrideItem{
						{
							Id:      "id-txt-1",
							Enabled: "1",
							Type:    "TXT",
							Domain:  "example.com",
							TxtData: "v=spf1 include:_spf.example.com ~all",
						},
						{
							Id:      "id-txt-2",
							Enabled: "1",
							Type:    "TXT",
							Domain:  "example.com",
							TxtData: "google-site-verification=abc123",
						},
					},
				},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			var body []endpoint.Endpoint
			Expect(json.Unmarshal(res.Body.Bytes(), &body)).To(Succeed())
			Expect(body).To(HaveLen(2))

			Expect(body[0].DNSName).To(Equal("example.com"))
			Expect(body[0].Targets).To(ConsistOf("v=spf1 include:_spf.example.com ~all"))
			Expect(body[0].RecordType).To(Equal("TXT"))

			Expect(body[1].DNSName).To(Equal("example.com"))
			Expect(body[1].Targets).To(ConsistOf("google-site-verification=abc123"))
			Expect(body[1].RecordType).To(Equal("TXT"))
		})
	})

	Context("POST", func() {
		It("should be able to validate the incoming headers", func() {
			c, res := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodPost, "/", nil))

			Expect(ctx.Respond(c, handler.HandleRecordsPost)).To(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusUnsupportedMediaType))
		})

		It("should be able to handle empty set of records", func() {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

			c, res := fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusNoContent))
		})

		When("deleting records", func() {
			It("should be able to handle A and AAAA records", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						Delete: []*endpoint.Endpoint{
							endpoint.NewEndpoint("example.com", endpoint.RecordTypeA, "192.168.1.1").
								WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id-A"),
							endpoint.NewEndpoint("example.com", endpoint.RecordTypeAAAA, "fd00::").
								WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id-AAAA"),
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				mocks.Client.EXPECT().UnboundDeleteHostOverride(mock.Anything, "id-A").Return(nil).Once()
				mocks.Client.EXPECT().UnboundDeleteHostOverride(mock.Anything, "id-AAAA").Return(nil).Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})

			It("should be able to handle TXT records", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						Delete: []*endpoint.Endpoint{
							endpoint.NewEndpoint("a-example.com", endpoint.RecordTypeTXT, "heritage=external-dns,external-dns/owner=test-cluster").
								WithSetIdentifier("e9ab3aba191cedd6e2c945ed0e976dbe72d0ca2676d1ac4a7e7907137abd4ee5").
								WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id-txt"),
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				// Mock the search call for registry TXT record matching
				mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(&unbound.SearchHostOverrideResponse{
					Rows: []unbound.SearchHostOverrideItem{
						{
							Id:      "id-txt1",
							Type:    endpoint.RecordTypeTXT,
							Domain:  "aaaa-example.com",
							TxtData: "heritage=external-dns,external-dns/owner=test-cluster,somethingelse",
						},
						{
							Id:      "id-txt2",
							Type:    endpoint.RecordTypeTXT,
							Domain:  "a-example.com",
							TxtData: "heritage=external-dns,external-dns/owner=test-cluster,somethingelse",
						},
						{
							Id:      "id-txt",
							Type:    endpoint.RecordTypeTXT,
							Domain:  "a-example.com",
							TxtData: "heritage=external-dns,external-dns/owner=test-cluster",
						},
					},
				}, nil).Once()

				mocks.Client.EXPECT().UnboundDeleteHostOverride(mock.Anything, "id-txt").Return(nil).Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})
		})

		When("updating records", func() {
			It("should be able to handle A and AAAA records", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						UpdateOld: []*endpoint.Endpoint{
							endpoint.NewEndpoint("example.com", endpoint.RecordTypeA, "192.168.1.1").
								WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id-A"),
							endpoint.NewEndpoint("example.com", endpoint.RecordTypeAAAA, "fd00::").
								WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id-AAAA"),
						},
						UpdateNew: []*endpoint.Endpoint{
							endpoint.NewEndpoint("example.com", endpoint.RecordTypeA, "192.168.1.1").
								WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id-A"),
							endpoint.NewEndpoint("example.com", endpoint.RecordTypeAAAA, "fd00::").
								WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id-AAAA"),
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				mocks.Client.EXPECT().
					UnboundUpdateHostOverride(mock.Anything, "id-A", &unbound.HostOverride{
						Enabled:  "1",
						Hostname: "example",
						Domain:   "com",
						Type:     "A",
						Server:   "192.168.1.1",
					}).
					Return(nil).
					Once()

				mocks.Client.EXPECT().
					UnboundUpdateHostOverride(mock.Anything, "id-AAAA", &unbound.HostOverride{
						Enabled:  "1",
						Hostname: "example",
						Domain:   "com",
						Type:     "AAAA",
						Server:   "fd00::",
					}).
					Return(nil).
					Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})

			It("should be able to handle TXT records", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						UpdateOld: []*endpoint.Endpoint{
							endpoint.NewEndpoint("a-example.com", endpoint.RecordTypeTXT, "heritage=external-dns,external-dns/owner=updated-cluster").
								WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id-txt"),
						},
						UpdateNew: []*endpoint.Endpoint{
							endpoint.NewEndpoint("a-example.com", endpoint.RecordTypeTXT, "heritage=external-dns,external-dns/owner=updated-cluster").
								WithProviderSpecific(provider.ProviderSpecificUUID.String(), "id-txt"),
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				mocks.Client.EXPECT().
					UnboundUpdateHostOverride(mock.Anything, "id-txt", &unbound.HostOverride{
						Enabled: "1",
						Domain:  "a-example.com",
						Type:    "TXT",
						TxtData: "heritage=external-dns,external-dns/owner=updated-cluster",
					}).
					Return(nil).
					Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})
		})

		When("creating records", func() {
			It("should be able to handle A and AAAA records with single target", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						Create: []*endpoint.Endpoint{
							endpoint.NewEndpoint("example.com", endpoint.RecordTypeA, "192.168.1.1"),
							endpoint.NewEndpoint("example.com", endpoint.RecordTypeAAAA, "fd00::"),
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled:  "1",
						Hostname: "example",
						Domain:   "com",
						Type:     "A",
						Server:   "192.168.1.1",
					}).
					Return("id-A", nil).
					Once()
				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled:  "1",
						Hostname: "example",
						Domain:   "com",
						Type:     "AAAA",
						Server:   "fd00::",
					}).
					Return("id-AAAA", nil).
					Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})

			It("should be able to handle A record with multiple targets", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						Create: []*endpoint.Endpoint{
							{
								DNSName:    "example.com",
								RecordType: endpoint.RecordTypeA,
								Targets:    []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
							},
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled:  "1",
						Hostname: "example",
						Domain:   "com",
						Type:     "A",
						Server:   "192.168.1.1",
					}).
					Return("id-A-1", nil).
					Once()
				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled:  "1",
						Hostname: "example",
						Domain:   "com",
						Type:     "A",
						Server:   "192.168.1.2",
					}).
					Return("id-A-2", nil).
					Once()
				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled:  "1",
						Hostname: "example",
						Domain:   "com",
						Type:     "A",
						Server:   "192.168.1.3",
					}).
					Return("id-A-3", nil).
					Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})

			It("should be able to handle TXT records with single target", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						Create: []*endpoint.Endpoint{
							endpoint.NewEndpoint("a-example.com", endpoint.RecordTypeTXT, "heritage=external-dns,external-dns/owner=test-cluster"),
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled: "1",
						Domain:  "a-example.com",
						Type:    "TXT",
						TxtData: "heritage=external-dns,external-dns/owner=test-cluster",
					}).
					Return("id-txt", nil).
					Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})

			It("should be able to handle TXT records with multiple targets", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						Create: []*endpoint.Endpoint{
							{
								DNSName:    "example.com",
								RecordType: endpoint.RecordTypeTXT,
								Targets:    []string{"v=spf1 include:_spf.example.com ~all", "google-site-verification=abc123"},
							},
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled: "1",
						Domain:  "example.com",
						Type:    "TXT",
						TxtData: "v=spf1 include:_spf.example.com ~all",
					}).
					Return("id-txt-1", nil).
					Once()
				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled: "1",
						Domain:  "example.com",
						Type:    "TXT",
						TxtData: "google-site-verification=abc123",
					}).
					Return("id-txt-2", nil).
					Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})

			It("should be able to create records with descriptions", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						Create: []*endpoint.Endpoint{
							endpoint.NewEndpoint("api.example.com", endpoint.RecordTypeA, "192.168.1.50").
								WithProviderSpecific(provider.ProviderSpecificDescription.String(), "Production API endpoint"),
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled:     "1",
						Hostname:    "api",
						Domain:      "example.com",
						Type:        "A",
						Server:      "192.168.1.50",
						Description: "Production API endpoint",
					}).
					Return("id-api", nil).
					Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})

			It("should be able to handle mixed A and TXT records (external-dns ownership pattern)", func() {
				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
						Create: []*endpoint.Endpoint{
							endpoint.NewEndpoint("example.com", endpoint.RecordTypeA, "192.168.1.1"),
							endpoint.NewEndpoint("a-example.com", endpoint.RecordTypeTXT, "heritage=external-dns,external-dns/owner=test-cluster"),
							endpoint.NewEndpoint("test.example.com", endpoint.RecordTypeAAAA, "fd00::1"),
							endpoint.NewEndpoint("aaaa-test.example.com", endpoint.RecordTypeTXT, "heritage=external-dns,external-dns/owner=test-cluster"),
						},
					})),
				)
				req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)

				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled:  "1",
						Hostname: "example",
						Domain:   "com",
						Type:     "A",
						Server:   "192.168.1.1",
					}).
					Return("id-A", nil).
					Once()
				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled: "1",
						Domain:  "a-example.com",
						Type:    "TXT",
						TxtData: "heritage=external-dns,external-dns/owner=test-cluster",
					}).
					Return("id-txt-a", nil).
					Once()
				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled:  "1",
						Hostname: "test",
						Domain:   "example.com",
						Type:     "AAAA",
						Server:   "fd00::1",
					}).
					Return("id-AAAA", nil).
					Once()
				mocks.Client.EXPECT().
					UnboundCreateHostOverride(mock.Anything, &unbound.HostOverride{
						Enabled: "1",
						Domain:  "aaaa-test.example.com",
						Type:    "TXT",
						TxtData: "heritage=external-dns,external-dns/owner=test-cluster",
					}).
					Return("id-txt-aaaa", nil).
					Once()

				c, res := fixtures.CreateEchoContext(nil, req)

				Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})
		})
	})
})
