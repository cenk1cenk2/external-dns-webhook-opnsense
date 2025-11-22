package probes_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/ctx"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("/readyz", func() {
	Context("GET", func() {
		It("should return http.StatusOK when not ready", func() {
			handler.IsReady = func() chan bool {
				c := make(chan bool, 1)

				c <- true

				return c
			}
			c, res := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodGet, "/", nil))

			Expect(ctx.Respond(c, handler.HandleReadyGet)).ToNot(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("should return http.StatusServiceUnavailable when not ready", func() {
			handler.IsReady = func() chan bool {
				c := make(chan bool, 1)

				c <- false

				return c
			}
			c, res := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodGet, "/", nil))

			Expect(ctx.Respond(c, handler.HandleReadyGet)).To(HaveOccurred())
			Expect(res.Code).To(Equal(http.StatusServiceUnavailable))
		})
	})
})
