package log

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	DebugLevel  = zap.DebugLevel
	InfoLevel   = zap.InfoLevel
	WarnLevel   = zap.WarnLevel
	ErrorLevel  = zap.ErrorLevel
	DPanicLevel = zap.DPanicLevel
	PanicLevel  = zap.PanicLevel
	FatalLevel  = zap.FatalLevel
)

func ConvertToLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "dpnic":
		return DPanicLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}
