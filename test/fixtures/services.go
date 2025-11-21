package fixtures

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
)

func NewTestLogger() *services.Logger {
	logger, err := services.NewLogger(&services.LoggerConfig{
		Level:   zapcore.DebugLevel.String(),
		Encoder: services.LogEncoderConsole,
	})

	Expect(err).ToNot(HaveOccurred())

	return logger
}
