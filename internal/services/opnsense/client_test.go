package opnsense_test

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services/opnsense"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Opnsense Client", func() {
	It("should create a new client", func(ctx SpecContext) {
		client, err := opnsense.NewClient(
			&opnsense.ClientSvc{
				Logger: fixtures.NewTestLogger(),
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
					Logger: fixtures.NewTestLogger(),
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
})
