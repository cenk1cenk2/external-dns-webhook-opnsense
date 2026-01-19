package services

import (
	"os"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	prettyconsole "github.com/thessem/zap-prettyconsole"
)

type LogEncoder string

const (
	LogEncoderConsole LogEncoder = "console"
	LogEncoderJson    LogEncoder = "json"
)

type Logger struct {
	*zap.Logger
}

type LoggerConfig struct {
	Level   string
	Encoder LogEncoder
}

type ZapLogger = *zap.Logger
type ZapSugaredLogger = *zap.SugaredLogger

func NewLogger(conf *LoggerConfig) (*Logger, error) {
	level, err := zap.ParseAtomicLevel(conf.Level)

	if err != nil {
		return nil, err
	}

	var encoder zapcore.Encoder

	switch conf.Encoder {
	case LogEncoderConsole:
		conf := prettyconsole.NewEncoderConfig()
		conf.CallerKey = "caller"
		conf.StacktraceKey = "stacktrace"
		conf.TimeKey = ""

		encoder = prettyconsole.NewEncoder(conf)
	default:
		encoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "logger",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		})
	}

	z := zap.New(
		zapcore.NewTee(
			zapcore.NewCore(
				encoder,
				zapcore.Lock(os.Stdout),
				zap.LevelEnablerFunc(func(l zapcore.Level) bool {
					return level.Level() <= l && l < zapcore.ErrorLevel
				}),
			),
			zapcore.NewCore(
				encoder,
				zapcore.Lock(os.Stderr),
				zap.LevelEnablerFunc(func(l zapcore.Level) bool {
					return level.Level() <= l && l >= zapcore.ErrorLevel
				}),
			),
		)).
		WithOptions(
			zap.AddStacktrace(zapcore.ErrorLevel),
		)

	defer func() {
		_ = z.Sync()
	}()

	logger := &Logger{
		z,
	}

	return logger, nil
}

func (l *Logger) WithEchoContext(c *echo.Context) ZapSugaredLogger {
	req := c.Request()
	res := c.Response()

	requestID := req.Header.Get(echo.HeaderXRequestID)
	if requestID == "" {
		requestID = res.Header().Get(echo.HeaderXRequestID)
	}

	return l.
		With(
			zap.String("protocol", req.Proto),
			zap.String("host", req.Host),
			zap.String("method", req.Method),
			zap.String("client_ip", c.RealIP()),
			zap.String("request_id", requestID),
			zap.String("path", req.RequestURI),
		).Sugar()
}

func (l *Logger) WithCaller() ZapSugaredLogger {
	return l.
		WithOptions(
			zap.AddCaller(),
		).
		Sugar()
}
