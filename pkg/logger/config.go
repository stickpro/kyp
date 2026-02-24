package logger

import (
	"fmt"
)

type Config struct {
	Format         LogFormat `yaml:"format" env:"FORMAT" default:"console" usage:"allows to set custom formatting" example:"json"`
	Level          LogLevel  `yaml:"level" env:"LEVEL" default:"info" usage:"allows to set custom logger level" example:"info"`
	ConsoleColored bool      `yaml:"console_colored" env:"CONSOLE_COLORED" default:"true" usage:"allows to set colored console output" example:"false"`
	Trace          LogLevel  `yaml:"trace" env:"TRACE" default:"fatal" usage:"allows to set custom trace level" example:"fatal"`
	WithCaller     bool      `yaml:"with_caller" env:"WITH_CALLER" default:"false" usage:"allows to show caller" example:"false"`
	WithStackTrace bool      `yaml:"with_stack_trace" env:"WITH_STACK_TRACE" default:"false" usage:"allows to show stack trace" example:"false"`
}

func (c *Config) Validate() error {
	if !c.Format.Valid() {
		return fmt.Errorf("invalid logger format: %s", c.Format)
	}

	if !c.Level.Valid() {
		return fmt.Errorf("invalid logger level: %s", c.Level)
	}

	if !c.Trace.Valid() {
		return fmt.Errorf("invalid logger trace level: %s", c.Trace)
	}

	return nil
}
