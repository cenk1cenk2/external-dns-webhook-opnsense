package webhook_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

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
)

var _ = Describe("records", func() {
	Context("GET", func() {
		It("should be able to validate the incoming headers", func() {
			c, res := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodGet, "/", nil))

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusNotAcceptable))
		})

		It("should be able to handle errors while fetching the records", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.AcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(nil, fmt.Errorf("")).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).To(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusUnprocessableEntity))
		})

		It("should be able to fetch the records on empty response", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.AcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(&unbound.SearchHostOverrideResponse{}, nil).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			var body []endpoint.Endpoint
			Expect(json.Unmarshal(res.Body.Bytes(), &body)).To(Succeed())
			Expect(body).To(BeEmpty())
		})

		It("should be able to fetch and convert all the records", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.AcceptedMedia)
			c, res := fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{
					Total:    1,
					RowCount: 1,
					Current:  1,
					Rows: []unbound.SearchHostOverrideItem{
						{
							Id:          "id",
							Enabled:     "1",
							Hostname:    "example",
							Domain:      "local",
							Type:        "A",
							Server:      "unbound",
							MXPriority:  "10",
							MXDomain:    "mail.local",
							Description: "test record",
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

			Expect(body[0].DNSName).To(Equal("example.local"))
			Expect(body[0].Targets).To(BeEquivalentTo([]string{"example.local"}))
			Expect(body[0].RecordType).To(Equal("A"))
			Expect(body[0].ProviderSpecific).To(ContainElements(
				endpoint.ProviderSpecificProperty{
					Name:  "opnsense.record.uuid",
					Value: "id",
				},
			))
		})

		It("should be able to fetch and convert records with ownership", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAccept, webhook.AcceptedMedia)
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
							Type:     "A",
							Hostname: "example",
							Domain:   "local",
							Server:   "127.0.0.1",
							Description: fixtures.MustJsonMarshal(provider.OwnershipRecord{
								Name:    "external-dns",
								Targets: []string{"example.local"},
							}),
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

			Expect(body[0].DNSName).To(Equal("example.local"))
			Expect(body[0].Targets).To(BeEquivalentTo([]string{"example.local"}))
			Expect(body[0].RecordType).To(Equal("A"))
			Expect(body[0].ProviderSpecific).To(ContainElements(
				endpoint.ProviderSpecificProperty{
					Name:  "opnsense.record.uuid",
					Value: "id",
				},
			))

			Expect(body[1].DNSName).To(Equal("external-dns"))
			Expect(body[1].Targets).To(BeEquivalentTo([]string{"example.local"}))
			Expect(body[1].RecordType).To(Equal("TXT"))
		})
	})
})
