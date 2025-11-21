package services_test

import (
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoggerService", func() {
	It("should be able to create a logger service", func(ctx SpecContext) {
		logger, err := services.NewLogger(&services.LoggerConfig{})

		Expect(err).ToNot(HaveOccurred())
		Expect(logger).ToNot(BeNil())
	})

	DescribeTable("should be able to create a logger service with level", func(ctx SpecContext, level string) {
		logger, err := services.NewLogger(&services.LoggerConfig{
			Level: level,
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(logger).ToNot(BeNil())
	},
		Entry("Fatal Level", zap.FatalLevel.String()),
		Entry("Error Level", zap.ErrorLevel.String()),
		Entry("Warn Level", zap.WarnLevel.String()),
		Entry("Info Level", zap.InfoLevel.String()),
		Entry("Debug Level", zap.DebugLevel.String()),
	)

	It("should not be able to create a logger service with invalid level", func(ctx SpecContext) {
		_, err := services.NewLogger(&services.LoggerConfig{
			Level: "invalid",
		})

		Expect(err).Should(HaveOccurred())
	})

	DescribeTable("should be able to create a logger service with log type", func(ctx SpecContext, t services.LogEncoder) {
		logger, err := services.NewLogger(&services.LoggerConfig{
			Encoder: t,
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(logger).ToNot(BeNil())
	},
		Entry("json", services.LogEncoderJson),
		Entry("console", services.LogEncoderConsole),
	)

	Context("service", func() {
		var logger *services.Logger

		BeforeEach(func() {
			var err error
			logger, err = services.NewLogger(&services.LoggerConfig{
				Level:   zapcore.DebugLevel.String(),
				Encoder: services.LogEncoderConsole,
			})

			Expect(err).ToNot(HaveOccurred())
		})

		It("should be able to create a sugared logger with echo context", func(ctx SpecContext) {
			c, _ := fixtures.CreateEchoContext(nil, httptest.NewRequest(http.MethodGet, "/test", nil))

			l := logger.WithEchoContext(c)

			Expect(func() {
				l.Infoln("test")
			}).
				ToNot(Panic())
		})

		It("should be able to create a sugared logger with service context", func(ctx SpecContext) {
			l := logger.WithCaller()

			Expect(func() {
				l.Infoln("test")
			}).
				ToNot(Panic())
		})
	})
})
