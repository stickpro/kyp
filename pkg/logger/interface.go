package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WriteSyncer is an io.Writer that can also flush any buffered data. Note
// that *os.File (and thus, os.Stderr and os.Stdout) implement WriteSyncer.
// Type alias.
type WriteSyncer = zapcore.WriteSyncer

// Sink defines the interface to write to and close logger destinations.
// Type alias.
type Sink = zap.Sink

// A SugaredLogger wraps the base Logger functionality in a slower, but less
// verbose, API. Any Logger can be converted to a SugaredLogger with its Sugar
// method.
// Type alias.
type sugaredLogger = zap.SugaredLogger

type ExtendedLogger interface {
	Logger
	Sync() error
	Std() *log.Logger
	Sugar() *sugaredLogger
}

// Logger common interface.
type Logger interface { //nolint:interfacebloat
	Debug(args ...any)
	Debugln(args ...any)
	Debugf(template string, args ...any)
	Debugw(msg string, keysAndValues ...any)

	Info(args ...any)
	Infoln(args ...any)
	Infof(template string, args ...any)
	Infow(msg string, keysAndValues ...any)

	Warn(args ...any)
	Warnln(args ...any)
	Warnf(template string, args ...any)
	Warnw(msg string, keysAndValues ...any)

	Error(args ...any)
	Errorln(args ...any)
	Errorf(template string, args ...any)
	Errorw(msg string, keysAndValues ...any)

	Fatal(args ...any)
	Fatalln(args ...any)
	Fatalf(template string, args ...any)
	Fatalw(msg string, keysAndValues ...any)

	Panic(args ...any)
	Panicln(args ...any)
	Panicf(template string, args ...any)
	Panicw(msg string, keysAndValues ...any)
}
