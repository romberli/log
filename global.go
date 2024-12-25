// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DefaultCallerSkip = 1
)

var (
	_globalL, _globalP, _ = NewLogger()
	_globalS              = _globalL.SugaredLogger
)

// L returns the global Logger, which can be reconfigured with ReplaceGlobals.
// It's safe for concurrent use.
func L() *Logger {
	return _globalL
}

// S returns the global SugaredLogger, which can be reconfigured with ReplaceGlobals.
// It's safe for concurrent use.
func S() *zap.SugaredLogger {
	return _globalS
}

// P returns the global properties, which can be reconfigured with ReplaceGlobals.
// It's safe for concurrent use.
func P() *ZapProperties {
	return _globalP
}

// ReplaceGlobals replaces global logger with given logger and properties
func ReplaceGlobals(logger *Logger, props *ZapProperties) {
	_globalL = logger.WithOptions(zap.AddCallerSkip(DefaultCallerSkip))
	_globalS = logger.Sugar()
	_globalP = props
}

// SetTimeFormat sets the time format of global logger
func SetTimeFormat(timeFormat string) {
	_globalL.SetTimeFormat(timeFormat)
	_globalP.Core.(*textIOCore).SetTimeFormat(timeFormat)
}

// SetSeperator sets the seperator of global logger
func SetSeperator(seperator string) {
	_globalL.SetSeperator(seperator)
	_globalP.Core.(*textIOCore).SetSeperator(seperator)
}

// SetDisableDoubleQuotes disables wrapping log content with double quotes of global logger
func SetDisableDoubleQuotes(disableDoubleQuotes bool) {
	_globalL.SetDisableDoubleQuotes(disableDoubleQuotes)
	_globalP.Core.(*textIOCore).SetDisableDoubleQuotes(disableDoubleQuotes)
}

// SetDisableEscape disables wrapping log content with double quotes of global logger
func SetDisableEscape(disableEscape bool) {
	_globalL.SetDisableEscape(disableEscape)
	_globalP.Core.(*textIOCore).SetDisableEscape(disableEscape)
}

// AddWriteSyncer add write syncer to multi write syncer, which allows to add a new way to write log message
func AddWriteSyncer(ws zapcore.WriteSyncer) {
	_globalL.AddWriteSyncer(ws)
}

// Clone clones global logger
func Clone() *Logger {
	return _globalL.Clone().WithOptions(zap.AddCallerSkip(-1))
}

// CloneAndAddWriteSyncer clones global logger and add specified write syncer to it
func CloneAndAddWriteSyncer(ws zapcore.WriteSyncer) *Logger {
	c := Clone()
	c.AddWriteSyncer(ws)
	return c
}

// CloneStdoutLogger clones global logger and add stdout write syncer to it
func CloneStdoutLogger() *Logger {
	return CloneAndAddWriteSyncer(NewStdoutWriteSyncer())
}

// Rotate rotates the log file of global logger
func Rotate() error {
	return L().Rotate()
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...zap.Field) {
	L().Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	S().Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	S().Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	S().Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	S().Errorf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(template string, args ...interface{}) {
	S().Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func Fatalf(template string, args ...interface{}) {
	S().Fatalf(template, args...)
}

// With creates a child logger and adds structured context to it.
// Fields added to the child don't affect the parent, and vice versa.
func With(fields ...zap.Field) *zap.Logger {
	return L().zapLogger.With(fields...)
}

// GetLevel gets the logging level.
func GetLevel() Level {
	return _globalP.Level.Level()
}

// SetLevel alters the logging level.
func SetLevel(l Level) {
	_globalP.Level.SetLevel(l)
}
