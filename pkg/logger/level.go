package logger

import (
	"go.uber.org/zap/zapcore"
)

type LogLevel string

// String returns log level as string.
func (l LogLevel) String() string { return string(l) }

// Valid checks if log level is valid.
func (l LogLevel) Valid() bool {
	switch l {
	case LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelFatal,
		LogLevelPanic:
		return true
	}
	return false
}

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

// safeLevel converts string representation into log level.
func safeLevel(level LogLevel) zapcore.Level {
	switch level {
	default:
		return zapcore.InfoLevel
	case LogLevelDebug:
		return zapcore.DebugLevel
	case LogLevelWarn:
		return zapcore.WarnLevel
	case LogLevelError:
		return zapcore.ErrorLevel
	case LogLevelPanic:
		return zapcore.PanicLevel
	case LogLevelFatal:
		return zapcore.FatalLevel
	}
}
