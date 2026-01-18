package webhook_test

import (
	"testing"

	h "github.com/cenk1cenk2/external-dns-webhook-opnsense/api/webhook"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/provider"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	mockservices "github.com/cenk1cenk2/external-dns-webhook-opnsense/test/mocks/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Webhook")
}

var handler *h.Handler
var mocks *MockServices

type MockServices struct {
	Client *mockservices.MockClientAdapter
}

var _ = BeforeEach(func(ctx SpecContext) {
	mocks = &MockServices{
		Client: mockservices.NewMockClientAdapter(GinkgoT()),
	}

	handler = h.NewHandler(&h.HandlerSvc{
		Log: fixtures.NewTestLogger(),
		Provider: &provider.Provider{
			Config:       provider.ProviderConfig{},
			Client:       mocks.Client,
			Log:          fixtures.NewTestLogger().Sugar(),
			DomainFilter: provider.NewDomainFilter(provider.DomainFilterConfig{}),
		},
	})
})
