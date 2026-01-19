package webhook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/browningluke/opnsense-go/pkg/unbound"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/webhook"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

var _ = Describe("E2E Flow", func() {
	Describe("Complete lifecycle - Multiple targets", func() {
		It("should handle the complete flow: create → reconcile → update → delete", func() {
			// Track created UUIDs to simulate OPNsense state
			created := make(map[string]*unbound.SearchHostOverrideItem)

			var (
				req *http.Request
				c   echo.Context
				res *httptest.ResponseRecorder
			)

			// ==========================================
			// SCENARIO 1: Initial Creation (Service with 2 IPs)
			// ==========================================
			By("Creating initial service with multiple targets")

			// Step 1: AdjustEndpoints - Source sends multi-target endpoint
			req = httptest.NewRequest(
				http.MethodPost,
				"/adjustendpoints",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "app.example.com",
						RecordType: endpoint.RecordTypeA,
						Targets:    []string{"10.0.0.1", "10.0.0.2"},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))

			adjusted := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(adjusted).To(HaveLen(2), "Should split into 2 endpoints")
			Expect(adjusted[0].SetIdentifier).ToNot(BeEmpty())
			Expect(adjusted[1].SetIdentifier).ToNot(BeEmpty())
			Expect(adjusted[0].SetIdentifier).ToNot(Equal(adjusted[1].SetIdentifier))

			// Step 2: Records - Check current state (empty initially)
			req = httptest.NewRequest(http.MethodGet, "/records", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{Rows: []unbound.SearchHostOverrideItem{}},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			current := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(current).To(BeEmpty(), "No existing records initially")

			// Step 3: ApplyChanges - Create both records
			req = httptest.NewRequest(
				http.MethodPost,
				"/records",
				strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{Create: adjusted})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			// Mock creates - OPNsense returns UUIDs
			for range adjusted {
				mocks.Client.EXPECT().UnboundCreateHostOverride(mock.Anything, mock.Anything).RunAndReturn(
					func(_ context.Context, h *unbound.HostOverride) (string, error) {
						id := uuid.NewString()
						created[id] = &unbound.SearchHostOverrideItem{
							Id: id, Enabled: h.Enabled, Hostname: h.Hostname,
							Domain: h.Domain, Type: h.Type, Server: h.Server,
						}
						return id, nil
					},
				).Once()
			}

			Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusNoContent))
			Expect(created).To(HaveLen(2), "Should have created 2 records in OPNsense")

			// ==========================================
			// SCENARIO 2: First Reconciliation (No changes)
			// ==========================================
			By("Reconciling with no changes")

			// Step 1: AdjustEndpoints - Same input
			req = httptest.NewRequest(
				http.MethodPost,
				"/adjustendpoints",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "app.example.com",
						RecordType: endpoint.RecordTypeA,
						Targets:    []string{"10.0.0.1", "10.0.0.2"},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			adjusted2 := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())

			// Verify SetIdentifiers are STABLE (same as before)
			Expect(adjusted2[0].SetIdentifier).To(Equal(adjusted[0].SetIdentifier))
			Expect(adjusted2[1].SetIdentifier).To(Equal(adjusted[1].SetIdentifier))

			// Step 2: Records - Fetch current state from OPNsense
			req = httptest.NewRequest(http.MethodGet, "/records", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			// Return created records from mock OPNsense
			rows := make([]unbound.SearchHostOverrideItem, 0, len(created))
			for _, r := range created {
				rows = append(rows, *r)
			}

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{Total: len(rows), RowCount: len(rows), Rows: rows},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			current = *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(current).To(HaveLen(2), "Should have 2 records from OPNsense")

			// Verify SetIdentifiers match between Desired and Current
			desiredKeys := map[string]bool{}
			for _, ep := range adjusted2 {
				desiredKeys[ep.SetIdentifier] = true
			}

			currentKeys := map[string]bool{}
			for _, ep := range current {
				currentKeys[ep.SetIdentifier] = true
			}

			Expect(currentKeys).To(HaveLen(2), "Should have 2 unique SetIdentifiers")
			Expect(desiredKeys).To(Equal(currentKeys), "SetIdentifiers should match - no changes needed")

			// Verify Current endpoints have UUIDs
			for _, ep := range current {
				id, exists := ep.GetProviderSpecificProperty(provider.ProviderSpecificUUID.String())
				Expect(exists).To(BeTrue(), "Current endpoints should have UUIDs")
				Expect(id).ToNot(BeEmpty())
			}

			// ==========================================
			// SCENARIO 3: Update (Change IP: 10.0.0.2 → 10.0.0.3)
			// ==========================================
			By("Updating service - changing one IP")

			// Step 1: AdjustEndpoints - Source sends updated targets
			req = httptest.NewRequest(
				http.MethodPost,
				"/adjustendpoints",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "app.example.com",
						RecordType: endpoint.RecordTypeA,
						Targets:    []string{"10.0.0.1", "10.0.0.3"}, // Changed!
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			adjusted3 := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())

			// First endpoint (10.0.0.1) should have SAME SetIdentifier
			// Second endpoint (10.0.0.3) should have NEW SetIdentifier
			Expect(adjusted3[0].SetIdentifier).To(Equal(adjusted[0].SetIdentifier), "Unchanged IP")
			Expect(adjusted3[1].SetIdentifier).ToNot(Equal(adjusted[1].SetIdentifier), "Changed IP")

			// Step 2: Records - Returns current state (still has 10.0.0.2)
			req = httptest.NewRequest(http.MethodGet, "/records", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{Total: len(rows), RowCount: len(rows), Rows: rows},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			current = *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())

			// Find the endpoint to delete (10.0.0.2)
			var endpointToDelete *endpoint.Endpoint
			var uuidToDelete string
			for _, ep := range current {
				if ep.Targets[0] == "10.0.0.2" {
					endpointToDelete = ep
					uuidToDelete, _ = ep.GetProviderSpecificProperty(provider.ProviderSpecificUUID.String())
					break
				}
			}
			Expect(endpointToDelete).ToNot(BeNil())
			Expect(uuidToDelete).ToNot(BeEmpty())

			// Find the endpoint to create (10.0.0.3)
			var endpointToCreate *endpoint.Endpoint
			for _, ep := range adjusted3 {
				if ep.Targets[0] == "10.0.0.3" {
					endpointToCreate = ep
					break
				}
			}
			Expect(endpointToCreate).ToNot(BeNil())

			// Step 3: ApplyChanges - Delete old, Create new
			req = httptest.NewRequest(
				http.MethodPost,
				"/records",
				strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{
					Delete: []*endpoint.Endpoint{endpointToDelete},
					Create: []*endpoint.Endpoint{endpointToCreate},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			// Mock delete - should use correct UUID
			mocks.Client.EXPECT().UnboundDeleteHostOverride(mock.Anything, uuidToDelete).Return(nil).Once()

			// Mock create - new record
			mocks.Client.EXPECT().UnboundCreateHostOverride(mock.Anything, mock.Anything).RunAndReturn(
				func(_ context.Context, h *unbound.HostOverride) (string, error) {
					Expect(h.Server).To(Equal("10.0.0.3"))
					id := uuid.NewString()
					created[id] = &unbound.SearchHostOverrideItem{
						Id: id, Enabled: h.Enabled, Hostname: h.Hostname,
						Domain: h.Domain, Type: h.Type, Server: h.Server,
					}
					return id, nil
				},
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// Clean up deleted record from our mock state
			delete(created, uuidToDelete)
			Expect(created).To(HaveLen(2), "Should still have 2 records (1 deleted, 1 created)")

			// ==========================================
			// SCENARIO 4: Delete All
			// ==========================================
			By("Deleting all records")

			// Step 1: Records - Get current state
			req = httptest.NewRequest(http.MethodGet, "/records", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			rows = make([]unbound.SearchHostOverrideItem, 0, len(created))
			for _, r := range created {
				rows = append(rows, *r)
			}

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{Total: len(rows), RowCount: len(rows), Rows: rows},
				nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			current = *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(current).To(HaveLen(2))

			// Step 2: ApplyChanges - Delete all
			req = httptest.NewRequest(
				http.MethodPost,
				"/records",
				strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{Delete: current})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			// Mock deletes - verify UUIDs are used correctly
			deletedUUIDs := make(map[string]bool)
			for _, ep := range current {
				id, _ := ep.GetProviderSpecificProperty(provider.ProviderSpecificUUID.String())
				deletedUUIDs[id] = true
				mocks.Client.EXPECT().UnboundDeleteHostOverride(mock.Anything, id).Return(nil).Once()
			}

			Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusNoContent))
			Expect(deletedUUIDs).To(HaveLen(2), "Should have deleted 2 records with correct UUIDs")
		})
	})

	Describe("Complete lifecycle - Single target", func() {
		It("should handle single-target endpoint without duplicates", func() {
			created := make(map[string]*unbound.SearchHostOverrideItem)

			var (
				req *http.Request
				c   echo.Context
				res *httptest.ResponseRecorder
			)

			// Step 1: AdjustEndpoints - Single target
			req = httptest.NewRequest(
				http.MethodPost,
				"/adjustendpoints",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{DNSName: "single.example.com", RecordType: endpoint.RecordTypeA, Targets: []string{"10.0.0.1"}},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			adjusted := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(adjusted).To(HaveLen(1))
			Expect(adjusted[0].SetIdentifier).ToNot(BeEmpty(), "Single target should get SetIdentifier")

			// Step 2: Create
			req = httptest.NewRequest(http.MethodGet, "/records", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, _ = fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{Rows: []unbound.SearchHostOverrideItem{}}, nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())

			req = httptest.NewRequest(
				http.MethodPost,
				"/records",
				strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{Create: adjusted})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, _ = fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundCreateHostOverride(mock.Anything, mock.Anything).RunAndReturn(
				func(_ context.Context, h *unbound.HostOverride) (string, error) {
					id := uuid.NewString()
					created[id] = &unbound.SearchHostOverrideItem{
						Id: id, Enabled: h.Enabled, Hostname: h.Hostname,
						Domain: h.Domain, Type: h.Type, Server: h.Server,
					}
					return id, nil
				},
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())

			// Step 3: Reconcile - Verify no duplicates
			req = httptest.NewRequest(
				http.MethodPost,
				"/adjustendpoints",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{DNSName: "single.example.com", RecordType: endpoint.RecordTypeA, Targets: []string{"10.0.0.1"}},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			adjusted2 := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(adjusted2[0].SetIdentifier).To(Equal(adjusted[0].SetIdentifier), "SetIdentifier should be stable")

			req = httptest.NewRequest(http.MethodGet, "/records", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			rows := make([]unbound.SearchHostOverrideItem, 0, len(created))
			for _, r := range created {
				rows = append(rows, *r)
			}
			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{Total: 1, RowCount: 1, Rows: rows}, nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			current := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(current).To(HaveLen(1))
			Expect(current[0].SetIdentifier).To(Equal(adjusted[0].SetIdentifier), "Current should match Desired")
		})
	})

	Describe("Complete lifecycle - TXT records", func() {
		It("should handle TXT records with multiple values", func() {
			created := make(map[string]*unbound.SearchHostOverrideItem)

			var (
				req *http.Request
				c   echo.Context
				res *httptest.ResponseRecorder
			)

			// Step 1: AdjustEndpoints - Multiple TXT values
			req = httptest.NewRequest(
				http.MethodPost,
				"/adjustendpoints",
				strings.NewReader(fixtures.MustJsonMarshal([]*endpoint.Endpoint{
					{
						DNSName:    "example.com",
						RecordType: endpoint.RecordTypeTXT,
						Targets:    []string{"v=spf1 include:_spf.example.com ~all", "google-site-verification=abc123"},
					},
				})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			Expect(ctx.Respond(c, handler.HandleAdjustEndpointsPost)).ToNot(HaveOccurred())
			adjusted := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(adjusted).To(HaveLen(2), "Should split into 2 TXT records")

			// Step 2: Create both TXT records
			req = httptest.NewRequest(http.MethodGet, "/records", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, _ = fixtures.CreateEchoContext(nil, req)

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{Rows: []unbound.SearchHostOverrideItem{}}, nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())

			req = httptest.NewRequest(
				http.MethodPost,
				"/records",
				strings.NewReader(fixtures.MustJsonMarshal(&plan.Changes{Create: adjusted})),
			)
			req.Header.Set(echo.HeaderContentType, webhook.ExternalDnsAcceptedMedia)
			c, _ = fixtures.CreateEchoContext(nil, req)

			for range adjusted {
				mocks.Client.EXPECT().UnboundCreateHostOverride(mock.Anything, mock.Anything).RunAndReturn(
					func(_ context.Context, h *unbound.HostOverride) (string, error) {
						id := uuid.NewString()
						created[id] = &unbound.SearchHostOverrideItem{
							Id: id, Enabled: h.Enabled, Domain: h.Domain, Type: h.Type, TxtData: h.TxtData,
						}
						return id, nil
					},
				).Once()
			}

			Expect(ctx.Respond(c, handler.HandleRecordsPost)).ToNot(HaveOccurred())
			Expect(created).To(HaveLen(2))

			// Step 3: Verify reconciliation works
			req = httptest.NewRequest(http.MethodGet, "/records", nil)
			req.Header.Set(echo.HeaderAccept, webhook.ExternalDnsAcceptedMedia)
			c, res = fixtures.CreateEchoContext(nil, req)

			rows := make([]unbound.SearchHostOverrideItem, 0, len(created))
			for _, r := range created {
				rows = append(rows, *r)
			}

			mocks.Client.EXPECT().UnboundSearchHostOverrides(mock.Anything).Return(
				&unbound.SearchHostOverrideResponse{Total: len(rows), RowCount: len(rows), Rows: rows}, nil,
			).Once()

			Expect(ctx.Respond(c, handler.HandleRecordsGet)).ToNot(HaveOccurred())
			current := *fixtures.MustJsonUnmarshal(&[]*endpoint.Endpoint{}, res.Body.Bytes())
			Expect(current).To(HaveLen(2))

			// Verify all have UUIDs
			for _, ep := range current {
				id, exists := ep.GetProviderSpecificProperty(provider.ProviderSpecificUUID.String())
				Expect(exists).To(BeTrue())
				Expect(id).ToNot(BeEmpty())
			}
		})
	})
})
