package log

import (
	"go.uber.org/zap"
)

var MyLogger *Logger

type Logger struct {
	zap.Logger
	SugaredLogger *zap.SugaredLogger
}

func NewMyLogger(logger *zap.Logger) *Logger {
	return &Logger{
		Logger:        *logger,
		SugaredLogger: logger.Sugar(),
	}
}

// Debugf uses fmt.Sprintf to log a templated message.
func (logger *Logger) Debugf(template string, args ...interface{}) {
	logger.SugaredLogger.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (logger *Logger) Infof(template string, args ...interface{}) {
	logger.SugaredLogger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (logger *Logger) Warnf(template string, args ...interface{}) {
	logger.SugaredLogger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (logger *Logger) Errorf(template string, args ...interface{}) {
	logger.SugaredLogger.Errorf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (logger *Logger) Panicf(template string, args ...interface{}) {
	logger.SugaredLogger.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (logger *Logger) Fatalf(template string, args ...interface{}) {
	logger.SugaredLogger.Fatalf(template, args...)
}
