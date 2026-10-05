package probes_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/probes"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/metrics"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("metrics", func() {
	Context("GET", func() {
		It("should expose the registered metrics in the prometheus format", func() {
			m := metrics.New()
			m.UnboundReconfigures.WithLabelValues("ok").Inc()

			a := probes.NewApi(&probes.ApiSvc{
				Logger:    fixtures.NewTestLogger(),
				Validator: services.NewValidator(),
				Metrics:   m,
			}, probes.ApiConfig{})

			res := httptest.NewRecorder()
			a.Echo.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/metrics", nil))

			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Body.String()).To(ContainSubstring(`opnsense_webhook_unbound_reconfigures_total{result="ok"} 1`))
			Expect(res.Body.String()).To(ContainSubstring("go_goroutines"))
		})
	})
})
