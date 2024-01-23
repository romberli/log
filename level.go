package log

import "strings"

type Level uint32

const (
	// set default caller skip which could modify caller section of log message
	DefaultCallerSkip = 1
	// PanicLevel level, the highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	case ErrorLevel:
		return "ERROR"
	case WarnLevel:
		return "WARN"
	case InfoLevel:
		return "INFO"
	case DebugLevel:
		return "DEBUG"
	case TraceLevel:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

func StringToLogLevel(level string) Level {
	switch strings.ToLower(level) {
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	case "error":
		return ErrorLevel
	case "warn", "warning":
		return WarnLevel
	case "info":
		return InfoLevel
	case "debug":
		return DebugLevel
	case "trace":
		return TraceLevel
	}

	return InfoLevel
}
