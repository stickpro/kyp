package logger_test

import (
	"fmt"
	"testing"

	"github.com/stickpro/kyp/pkg/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	l := logger.New(
		logger.WithAppName("some app name"),
		logger.WithAppVersion("v0.1.0"),
		logger.WithLogLevel(logger.LogLevelDebug),
		logger.WithCaller(true),
		logger.WithStackTrace(true),
		logger.WithZapOption(zap.Hooks(func(entry zapcore.Entry) error {
			fmt.Println("hook")
			return nil
		})),
	)

	l.Info("Hello world")
}

func Test_LoggerWith(t *testing.T) {
	l := logger.New(
		logger.WithAppName("test app name"),
		logger.WithLogLevel(logger.LogLevelDebug),
	)

	l = logger.With(l, "key", "value")

	l = logger.With(l, "key2", "value2")

	l = logger.With(l, "key3", "value3")

	l.Infow("some test value", "numbers", 1234)
}
