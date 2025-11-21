package probes_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("/healthz", func() {
	Context("GET", func() {
		It("should return http.StatusOK when ready", func() {
			c, res := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodGet, "/", nil))

			Expect(ctx.RespondWithContext(c, handler.HandleHealthGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
		})
	})
})
