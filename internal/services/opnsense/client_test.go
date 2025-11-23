package opnsense_test

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/unbound"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/opnsense"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Opnsense Client", func() {
	It("should create a new client", func(ctx SpecContext) {
		client, err := opnsense.NewClient(
			&opnsense.OpnsenseClientSvc{
				Logger: fixtures.NewTestLogger(),
			},
			opnsense.OpnsenseClientConfig{
				Options: api.Options{
					Uri:       "opnsense.invalid",
					APIKey:    "testkey",
					APISecret: "testsecret",
				},
				DryRun: false,
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
			client = &opnsense.Client{
				Log:    fixtures.NewTestLogger().Sugar(),
				Config: opnsense.OpnsenseClientConfig{DryRun: true},
			}
		})

		It("should not modify anything on create", func(ctx SpecContext) {
			uuid, err := client.UnboundCreateHostOverride(ctx, &unbound.HostOverride{})

			Expect(err).ToNot(HaveOccurred())
			Expect(uuid).To(Equal(""))
		})

		It("should not modify anything on update", func(ctx SpecContext) {
			err := client.UnboundUpdateHostOverride(ctx, "", &unbound.HostOverride{})

			Expect(err).ToNot(HaveOccurred())
		})

		It("should not modify anything on delete", func(ctx SpecContext) {
			err := client.UnboundDeleteHostOverride(ctx, "")

			Expect(err).ToNot(HaveOccurred())
		})
	})
})
