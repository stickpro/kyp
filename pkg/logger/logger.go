package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	config Config

	appName    string
	appVersion string

	zapConfig zapcore.EncoderConfig
	options   []zap.Option

	*sugaredLogger
}

// Sugar returns zap.SugaredLogger
func (l *logger) Sugar() *sugaredLogger { return l.sugaredLogger }

// Std returns standard library log.Logger
func (l *logger) Std() *log.Logger { return zap.NewStdLog(l.Desugar()) }

// Logger returns logger instance
func (l *logger) LoggerInstance() *logger {
	return l
}

// Default returns default logger instance
func Default() Logger {
	return DefaultExtended()
}

// Default returns default extended logger instance
func DefaultExtended() ExtendedLogger {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(zapcore.DebugLevel)

	encoderCfg := zap.NewProductionEncoderConfig()

	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	l := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			atom,
		),
		zap.AddCaller(),
	)

	return &logger{sugaredLogger: l.Sugar()}
}

// New - init new logger with options
func New(opts ...Option) Logger {
	return initLogger(opts...)
}

// NewExtended - init new extended logger with options
func NewExtended(opts ...Option) ExtendedLogger {
	return initLogger(opts...)
}

// With allows to provide zap.SugaredLogger as common interface.
func With(l Logger, args ...any) Logger {
	lgIface, ok := l.(interface{ LoggerInstance() *logger })
	if !ok {
		return l
	}

	lg := lgIface.LoggerInstance()

	return &logger{
		lg.config,
		lg.appName,
		lg.appVersion,
		lg.zapConfig,
		lg.options,
		lg.With(args...),
	}
}

// WithExtended allows to provide zap.SugaredLogger as common interface.
func WithExtended(l ExtendedLogger, args ...any) ExtendedLogger {
	lgIface, ok := l.(interface{ LoggerInstance() *logger })
	if !ok {
		return l
	}

	lg := lgIface.LoggerInstance()

	return &logger{
		lg.config,
		lg.appName,
		lg.appVersion,
		lg.zapConfig,
		lg.options,
		lg.With(args...),
	}
}

func initLogger(opts ...Option) *logger {
	var l logger
	l.zapConfig = zap.NewProductionEncoderConfig()

	for _, o := range opts {
		o(&l)
	}

	logLevel := safeLevel(l.config.Level)
	logTrace := safeLevel(l.config.Trace)

	l.zapConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	encoder := zapcore.NewJSONEncoder(l.zapConfig)
	if l.config.Format == LoggerFormatConsole {
		if l.config.ConsoleColored {
			l.zapConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(l.zapConfig)
	}

	buildOpts := l.options
	if l.config.WithCaller {
		buildOpts = append(buildOpts, zap.AddCaller())
	}

	if l.config.WithStackTrace {
		buildOpts = append(buildOpts, zap.AddStacktrace(logTrace))
	}

	atom := zap.NewAtomicLevel()
	atom.SetLevel(logLevel)

	zapLogger := zap.New(
		zapcore.NewCore(
			encoder,
			zapcore.Lock(os.Stdout),
			atom,
		),
		buildOpts...,
	)

	if l.appName != "" {
		zapLogger = zapLogger.With(zap.String("app", l.appName))
	}

	if l.appVersion != "" {
		zapLogger = zapLogger.With(zap.String("version", l.appVersion))
	}

	l.sugaredLogger = zapLogger.Sugar()

	return &l
}
