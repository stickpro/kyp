package logger

// A LogFormat is a string that represents the log format.
type LogFormat string

const (
	LoggerFormatConsole LogFormat = "console"
	LoggerFormatJSON    LogFormat = "json"
)

// Valid checks if log format is valid.
func (f LogFormat) Valid() bool {
	switch f {
	case LoggerFormatConsole, LoggerFormatJSON:
		return true
	}
	return false
}
