package logger

import (
	"strings"

	"go.uber.org/zap"
)

type Option func(*logger)

func WithConfig(v Config) Option {
	return func(l *logger) { l.config = v }
}

func WithLogLevel(v LogLevel) Option {
	return func(l *logger) { l.config.Level = LogLevel(strings.ToLower(v.String())) }
}

func WithLogFormat(v LogFormat) Option {
	return func(l *logger) {
		switch v {
		case LoggerFormatConsole, LoggerFormatJSON:
			l.config.Format = v
		}
	}
}

// WithConsoleColored allows to set colored console output.
func WithConsoleColored(v bool) Option {
	return func(l *logger) {
		l.config.ConsoleColored = v
	}
}

// WithAppName allows to set application name to logger fields.
func WithAppName(v string) Option {
	return func(l *logger) { l.appName = v }
}

// WithAppVersion allows to set application version to logger fields.
func WithAppVersion(v string) Option {
	return func(l *logger) { l.appVersion = v }
}

func WithTimeKey(v string) Option {
	return func(l *logger) { l.zapConfig.TimeKey = v }
}

func WithCaller(v bool) Option {
	return func(l *logger) { l.config.WithCaller = v }
}

func WithStackTrace(v bool) Option {
	return func(l *logger) { l.config.WithStackTrace = v }
}

// WithZapOption allows to set zap.Option.
func WithZapOption(v zap.Option) Option {
	return func(l *logger) { l.options = append(l.options, v) }
}
