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

type Level uint32

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
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

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debugf(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Infof(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warnf(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Errorf(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panicf(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatalf(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debug(template string, args ...interface{}) {
	S().Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Info(template string, args ...interface{}) {
	S().Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warn(template string, args ...interface{}) {
	S().Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Error(template string, args ...interface{}) {
	S().Errorf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panic(template string, args ...interface{}) {
	S().Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func Fatal(template string, args ...interface{}) {
	S().Fatalf(template, args...)
}

// With creates a child logger and adds structured context to it.
// Fields added to the child don't affect the parent, and vice versa.
func With(fields ...zap.Field) *zap.Logger {
	return L().WithOptions(zap.AddCallerSkip(1)).With(fields...)
}

// SetLevel alters the logging level.
func SetLevel(l zapcore.Level) {
	_globalP.Level.SetLevel(l)
}

// GetLevel gets the logging level.
func GetLevel() zapcore.Level {
	return _globalP.Level.Level()
}
