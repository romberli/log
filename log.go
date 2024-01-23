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
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pingcap/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const DefaultOutput = "stdout"

var (
	MyLogger *Logger
	MyProps  *ZapProperties
)

// textFormatter is for compatibility with ngaut/log
type textFormatter struct {
	DisableTimestamp bool
	EnableEntryOrder bool
}

// Format implements logrus.Formatter
func (f *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var (
		err error
		b   *bytes.Buffer
	)
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	if !f.DisableTimestamp {
		_, err = fmt.Fprintf(b, "%s ", entry.Time.Format(defaultLogTimeFormat))
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	if file, ok := entry.Data["file"]; ok {
		_, err = fmt.Fprintf(b, "%s:%v:", file, entry.Data["line"])
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	_, err = fmt.Fprintf(b, " [%s] %s", entry.Level.String(), entry.Message)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if f.EnableEntryOrder {
		keys := make([]string, 0, len(entry.Data))
		for k := range entry.Data {
			if k != "file" && k != "line" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			_, err = fmt.Fprintf(b, " %v=%v", k, entry.Data[k])
			if err != nil {
				return nil, errors.Trace(err)
			}
		}
	} else {
		for k, v := range entry.Data {
			if k != "file" && k != "line" {
				_, err = fmt.Fprintf(b, " %v=%v", k, v)
				if err != nil {
					return nil, errors.Trace(err)
				}
			}
		}
	}

	err = b.WriteByte('\n')
	if err != nil {
		return nil, errors.Trace(err)
	}

	return b.Bytes(), nil
}

func StringToLogFormatter(format string, disableTimestamp bool) logrus.Formatter {
	switch strings.ToLower(format) {
	case "text":
		return &textFormatter{
			DisableTimestamp: disableTimestamp,
		}
	default:
		return &textFormatter{}
	}
}

// InitLumberjackLoggerWithFileLogConfig initializes file based logging options.
func InitLumberjackLoggerWithFileLogConfig(cfg *FileLogConfig) (*Writer, error) {
	st, err := os.Stat(cfg.FileName)
	if err == nil {
		if st.IsDir() {
			return nil, errors.New("can't use directory as log file name")
		}
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = DefaultLogMaxSize
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = DefaultLogMaxBackups
	}
	if cfg.MaxDays <= 0 {
		cfg.MaxDays = DefaultLogMaxDays
	}

	// use lumberjack to rotate log file
	return &Writer{
		Filename:   cfg.FileName,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxDays,
		LocalTime:  true,
		Options:    cfg.Options,
	}, nil
}

// NewLogger returns a logger which will write log message to stdout with default log level and format
func NewLogger() (*Logger, *ZapProperties, error) {
	var err error

	MyLogger, MyProps, err = NewStdoutLogger(DefaultLogLevel, defaultLogTimeFormat)

	return MyLogger, MyProps, err
}

// NewStdoutLogger returns a logger which will write log message to stdout
func NewStdoutLogger(level, format string) (*Logger, *ZapProperties, error) {
	cfg := &Config{
		Level:  level,
		Format: format,
		File:   FileLogConfig{},
	}

	multiWriteSyncer := NewMultiWriteSyncer(NewStdoutWriteSyncer())
	myZapLogger, myProps, err := InitZapLoggerWithWriteSyncer(
		cfg, multiWriteSyncer, zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Development(),
	)
	if err != nil {
		return nil, nil, err
	}

	myLogger := NewMyLogger(myZapLogger)

	return myLogger, myProps, err
}

// InitLoggerWithConfig initializes a zap logger with config.
func InitLoggerWithConfig(cfg *Config) (*Logger, *ZapProperties, error) {
	var (
		err              error
		writer           *Writer
		output           zapcore.WriteSyncer
		multiWriteSyncer zapcore.WriteSyncer
		zapLogger        *zap.Logger
	)

	if len(cfg.File.FileName) > 0 {
		writer, err = InitLumberjackLoggerWithFileLogConfig(&cfg.File)
		if err != nil {
			return nil, nil, err
		}

		output = NewWriteSyncer(writer)
	} else {
		output = NewStdoutWriteSyncer()
	}

	multiWriteSyncer = NewMultiWriteSyncer(output)
	zapLogger, MyProps, err = InitZapLoggerWithWriteSyncer(
		cfg, multiWriteSyncer, zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Development(),
	)
	if err != nil {
		return nil, nil, err
	}

	MyLogger = NewMyLogger(zapLogger)
	ReplaceGlobals(MyLogger, MyProps)

	return MyLogger, MyProps, nil
}

// InitStdoutLogger initiates a stdout logger with given level and format
func InitStdoutLogger(level, format string) (*Logger, *ZapProperties, error) {
	cfg := NewConfigWithStdout(level, format)

	return InitLoggerWithConfig(cfg)
}

// InitStdoutLogger initiates a stdout logger with given level and format
func InitStdoutLoggerWithDefault() (*Logger, *ZapProperties, error) {
	return InitStdoutLogger(DefaultLogLevel, DefaultLogFormat)
}

// InitFileLogger initiates a file logger with given options
func InitFileLogger(fileName, level, format string, maxSize, maxDays, maxBackups int, options ...Option) (*Logger, *ZapProperties, error) {
	cfg, err := NewConfigWithFileLog(fileName, level, format, maxSize, maxDays, maxBackups, options...)
	if err != nil {
		return nil, nil, err
	}

	return InitLoggerWithConfig(cfg)
}

// InitFileLoggerWithDefaultConfig initiates logger with default options
func InitFileLoggerWithDefault(fileName string) (*Logger, *ZapProperties, error) {
	cfg, err := NewConfigWithFileLog(fileName, DefaultLogLevel, DefaultLogFormat, DefaultLogMaxSize, DefaultLogMaxDays, DefaultLogMaxBackups, nil)
	if err != nil {
		return nil, nil, err
	}

	return InitLoggerWithConfig(cfg)
}

// InitZapLoggerWithWriteSyncer initializes a zap logger with specified  write syncer.
func InitZapLoggerWithWriteSyncer(cfg *Config, output zapcore.WriteSyncer, opts ...zap.Option) (*zap.Logger, *ZapProperties, error) {
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	core := NewTextCore(newZapTextEncoder(cfg).(*textEncoder), output, level)
	opts = append(cfg.buildOptions(output), opts...)
	lg := zap.New(core, opts...)
	r := &ZapProperties{
		Core:   core,
		Syncer: output,
		Level:  level,
	}

	return lg, r, nil
}

// CloneLogger returns a fresh new logger with same options
func CloneLogger(logger *Logger) *Logger {
	return logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core { return c.With([]zapcore.Field{}) }))
}

// init initiate MyLogger when this package is imported
func init() {
	MyLogger, MyProps, _ = NewLogger()
	ReplaceGlobals(MyLogger, MyProps)
}
