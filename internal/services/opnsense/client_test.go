package opnsense_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/metrics"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/opnsense"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

var _ = Describe("Opnsense Client", func() {
	It("should create a new client", func(ctx SpecContext) {
		client, err := opnsense.NewClient(
			&opnsense.ClientSvc{
				Logger:  fixtures.NewTestLogger(),
				Metrics: metrics.New(),
			},
			opnsense.ClientConfig{
				Uri:           "opnsense.invalid",
				APIKey:        "testkey",
				APISecret:     "testsecret",
				AllowInsecure: false,
				DryRun:        false,
			},
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(client).ToNot(BeNil())
	})

	Context("dry run client", func() {
		var (
			client *opnsense.Client
		)

		BeforeEach(func() {
			c, err := opnsense.NewClient(
				&opnsense.ClientSvc{
					Logger:  fixtures.NewTestLogger(),
					Metrics: metrics.New(),
				},
				opnsense.ClientConfig{
					Uri:           "opnsense.invalid",
					APIKey:        "testkey",
					APISecret:     "testsecret",
					AllowInsecure: false,
					DryRun:        true,
				},
			)
			Expect(err).ToNot(HaveOccurred())
			client = c
		})

		It("should not modify anything on create", func(ctx SpecContext) {
			uuid, err := client.UnboundCreateHostOverride(ctx, &opnsense.UnboundHostOverride{})

			Expect(err).ToNot(HaveOccurred())
			Expect(uuid).To(Equal(""))
		})

		It("should not modify anything on update", func(ctx SpecContext) {
			err := client.UnboundUpdateHostOverride(ctx, "", &opnsense.UnboundHostOverride{})

			Expect(err).ToNot(HaveOccurred())
		})

		It("should not modify anything on delete", func(ctx SpecContext) {
			err := client.UnboundDeleteHostOverride(ctx, "")

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("request timeout", func() {
		It("should give up on a slow server within the retry budget", func(ctx SpecContext) {
			release := make(chan struct{})
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				<-release
			}))
			DeferCleanup(srv.Close)
			DeferCleanup(func() { close(release) })

			client, err := opnsense.NewClient(
				&opnsense.ClientSvc{
					Logger:  fixtures.NewTestLogger(),
					Metrics: metrics.New(),
				},
				opnsense.ClientConfig{
					Uri:        srv.URL,
					APIKey:     "testkey",
					APISecret:  "testsecret",
					Timeout:    200 * time.Millisecond,
					MaxRetries: 1,
					MinBackoff: 50 * time.Millisecond,
					MaxBackoff: 50 * time.Millisecond,
				},
			)
			Expect(err).ToNot(HaveOccurred())

			start := time.Now()
			err = client.CheckUnboundService(ctx)

			Expect(err).To(HaveOccurred())
			Expect(time.Since(start)).To(BeNumerically("<", 1*time.Second))
		})
	})

	Context("metrics", func() {
		var (
			m      *metrics.Metrics
			status *int
			calls  *int
			client *opnsense.Client
		)

		BeforeEach(func() {
			m = metrics.New()
			code, count := http.StatusOK, 0
			status, calls = &code, &count

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				*calls++
				w.WriteHeader(*status)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			DeferCleanup(srv.Close)

			c, err := opnsense.NewClient(
				&opnsense.ClientSvc{
					Logger:  fixtures.NewTestLogger(),
					Metrics: m,
				},
				opnsense.ClientConfig{
					Uri:        srv.URL,
					APIKey:     "testkey",
					APISecret:  "testsecret",
					Timeout:    time.Second,
					MaxRetries: 1,
					MinBackoff: time.Millisecond,
					MaxBackoff: time.Millisecond,
				},
			)
			Expect(err).ToNot(HaveOccurred())
			client = c
		})

		It("should count successful reconfigures and requests without leaking identifiers", func(ctx SpecContext) {
			Expect(client.ReconfigureService(ctx)).To(Succeed())
			Expect(client.UnboundDeleteHostOverride(ctx, "7c1d")).To(HaveOccurred())

			Expect(testutil.ToFloat64(m.UnboundReconfigures.WithLabelValues("ok"))).To(Equal(1.0))
			Expect(testutil.ToFloat64(m.ClientRequests.WithLabelValues("POST", "/unbound/service/reconfigure", "200"))).To(Equal(1.0))
			Expect(testutil.ToFloat64(m.ClientRequests.WithLabelValues("POST", "/unbound/settings/delHostOverride", "200"))).To(Equal(1.0))
			Expect(testutil.CollectAndCount(m.ClientRequestDuration)).To(Equal(2))
		})

		It("should count failed reconfigures and retries on server errors", func(ctx SpecContext) {
			*status = http.StatusInternalServerError

			Expect(client.ReconfigureService(ctx)).To(HaveOccurred())

			Expect(*calls).To(Equal(2))
			Expect(testutil.ToFloat64(m.UnboundReconfigures.WithLabelValues("error"))).To(Equal(1.0))
			Expect(testutil.ToFloat64(m.ClientRetries)).To(Equal(1.0))
		})

		It("should label transport failures as errors", func(ctx SpecContext) {
			c, err := opnsense.NewClient(
				&opnsense.ClientSvc{
					Logger:  fixtures.NewTestLogger(),
					Metrics: m,
				},
				opnsense.ClientConfig{
					Uri:        "http://127.0.0.1:1",
					Timeout:    100 * time.Millisecond,
					MinBackoff: time.Millisecond,
					MaxBackoff: time.Millisecond,
				},
			)
			Expect(err).ToNot(HaveOccurred())

			Expect(c.CheckUnboundService(ctx)).To(HaveOccurred())
			Expect(testutil.ToFloat64(m.ClientRequests.WithLabelValues("POST", "/core/service/search", "error"))).To(Equal(1.0))
		})
	})
})
