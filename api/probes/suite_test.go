package probes_test

import (
	"testing"

	h "github.com/cenk1cenk2/external-dns-webhook-opnsense/api/probes"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Health Probes")
}

var handler *h.Handler

var _ = BeforeEach(func(ctx SpecContext) {
	handler = h.NewHandler(&h.HandlerSvc{
		Log: fixtures.NewTestLogger(),
	})
})
