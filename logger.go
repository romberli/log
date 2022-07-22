package log

import (
	"github.com/pingcap/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger and sugared logger
type Logger struct {
	zapLogger     *zap.Logger
	SugaredLogger *zap.SugaredLogger
}

// NewMyLogger returns *Logger
func NewMyLogger(logger *zap.Logger) *Logger {
	return &Logger{
		zapLogger:     logger.WithOptions(zap.AddCallerSkip(DefaultCallerSkip)),
		SugaredLogger: logger.WithOptions(zap.AddCallerSkip(DefaultCallerSkip)).Sugar(),
	}
}

func (logger *Logger) Rotate() error {
	core, ok := logger.zapLogger.Core().(*textIOCore)
	if ok {
		ws, ok := core.GetWriterSyncer().(*WriteSyncer)
		if ok {
			w, ok := ws.GetWriter().(*Writer)
			if ok {
				return w.Rotate()
			}
		}
	}

	return errors.New("failed to rotate log file, make sure use lumberjack writer as the writer")
}

// Clone clones logger and returns the new one
func (logger *Logger) Clone() *Logger {
	return CloneLogger(logger)
}

// SetDisableDoubleQuotes disables wrapping log content with double quotes
func (logger *Logger) SetDisableDoubleQuotes(disableDoubleQuotes bool) {
	logger.zapLogger.Core().(*textIOCore).SetDisableDoubleQuotes(disableDoubleQuotes)
}

// SetDisableEscape disables escaping special characters of log content like \n,\r...
func (logger *Logger) SetDisableEscape(disableEscape bool) {
	logger.zapLogger.Core().(*textIOCore).SetDisableEscape(disableEscape)
}

// AddWriteSyncer adds write syncer to multi write syncer, which allows to add a new way to write log message
func (logger *Logger) AddWriteSyncer(ws zapcore.WriteSyncer) {
	logger.zapLogger.Core().(*textIOCore).AddWriteSyncer(ws)
}

// CloneAndAddWriteSyncer adds write syncer to multi write syncer, which allows to add a new way to write log message
func (logger *Logger) CloneAndAddWriteSyncer(ws zapcore.WriteSyncer) *Logger {
	c := logger.Clone()
	c.AddWriteSyncer(ws)
	return c
}

// WithOptions returns a new *Logger with specified options
func (logger *Logger) WithOptions(opts ...zap.Option) *Logger {
	return &Logger{
		zapLogger:     logger.zapLogger.WithOptions(opts...),
		SugaredLogger: logger.zapLogger.WithOptions(opts...).Sugar(),
	}
}

// Sugar returns a new sugared logger
func (logger *Logger) Sugar() *zap.SugaredLogger {
	return logger.zapLogger.Sugar()
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (logger *Logger) Debug(msg string, fields ...zap.Field) {
	logger.zapLogger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (logger *Logger) Info(msg string, fields ...zap.Field) {
	logger.zapLogger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (logger *Logger) Warn(msg string, fields ...zap.Field) {
	logger.zapLogger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (logger *Logger) Error(msg string, fields ...zap.Field) {
	logger.zapLogger.Error(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (logger *Logger) Panic(msg string, fields ...zap.Field) {
	logger.zapLogger.Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (logger *Logger) Fatal(msg string, fields ...zap.Field) {
	logger.zapLogger.Fatal(msg, fields...)
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
