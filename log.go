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
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DefaultOutput = "stdout"
)

var (
	MyLogger    *Logger
	MyZapLogger *zap.Logger
	MyProps     *ZapProperties
)

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
			return nil, err
		}
	}
	if file, ok := entry.Data["file"]; ok {
		_, err = fmt.Fprintf(b, "%s:%v:", file, entry.Data["line"])
		if err != nil {
			return nil, err
		}
	}
	_, err = fmt.Fprintf(b, " [%s] %s", entry.Level.String(), entry.Message)
	if err != nil {
		return nil, err
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
				return nil, err
			}
		}
	} else {
		for k, v := range entry.Data {
			if k != "file" && k != "line" {
				_, err = fmt.Fprintf(b, " %v=%v", k, v)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	err = b.WriteByte('\n')
	if err != nil {
		return nil, err
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

// initFileLog initializes file based logging options.
func initFileLog(cfg *FileLogConfig) (*lumberjack.Logger, error) {
	if st, err := os.Stat(cfg.FileName); err == nil {
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
	return &lumberjack.Logger{
		Filename:   cfg.FileName,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxDays,
		LocalTime:  true,
	}, nil
}

// newLogger returns a logger
func newLogger() (*Logger, *ZapProperties, error) {
	var (
		err       error
		cfg       *Config
		stdOut    zapcore.WriteSyncer
		closeFunc func()
	)

	cfg = &Config{
		Level:  DefaultLogLevel,
		Format: DefaultLogFormat,
		File:   FileLogConfig{}}

	stdOut, closeFunc, err = zap.Open([]string{DefaultOutput}...)
	if err != nil {
		if closeFunc != nil {
			closeFunc()
		}

		return nil, nil, errors.Trace(err)
	}

	MyZapLogger, MyProps, err = InitLoggerWithWriteSyncer(
		cfg, stdOut, zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCallerSkip(DefaultCallerSkip),
		zap.Development(),
	)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	MyLogger = NewMyLogger(MyZapLogger)

	return MyLogger, MyProps, err
}

// InitLoggerWithConfig initializes a zap logger with config.
func InitLoggerWithConfig(cfg *Config) (*Logger, *ZapProperties, error) {
	var (
		err       error
		lg        *lumberjack.Logger
		stdOut    zapcore.WriteSyncer
		closeFunc func()
		output    zapcore.WriteSyncer
	)

	if len(cfg.File.FileName) > 0 {
		lg, err = initFileLog(&cfg.File)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}

		output = zapcore.AddSync(lg)
	} else {
		stdOut, closeFunc, err = zap.Open([]string{DefaultOutput}...)
		if err != nil {
			if closeFunc != nil {
				closeFunc()
			}

			return nil, nil, errors.Trace(err)
		}

		output = stdOut
	}

	zapLogger, MyProps, err := InitLoggerWithWriteSyncer(
		cfg, output, zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCaller(),
		zap.Development(),
	)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	MyLogger = NewMyLogger(zapLogger)
	ReplaceGlobals(MyLogger, MyProps)

	return MyLogger, MyProps, nil
}

// InitLogger initiate logger with given options
func InitLogger(fileName, level, format string, maxSize, maxDays, maxBackups int) (*Logger, *ZapProperties, error) {
	logConfig, err := NewConfigWithFileLog(fileName, level, format, maxSize, maxDays, maxBackups)
	if err != nil {
		fmt.Printf("got error when creating log config.\n%s", err.Error())
	}

	return InitLoggerWithConfig(logConfig)
}

// InitLogger initiate logger with default options
func InitLoggerWithDefaultConfig(fileName string) (*Logger, *ZapProperties, error) {
	logConfig, err := NewConfigWithFileLog(fileName, DefaultLogLevel, DefaultLogFormat, DefaultLogMaxSize, DefaultLogMaxDays, DefaultLogMaxBackups)
	if err != nil {
		fmt.Printf("got error when creating log config.\n%s", err.Error())
	}

	return InitLoggerWithConfig(logConfig)
}

// InitLoggerWithWriteSyncer initializes a zap logger with specified  write syncer.
func InitLoggerWithWriteSyncer(cfg *Config, output zapcore.WriteSyncer, opts ...zap.Option) (*zap.Logger, *ZapProperties, error) {
	level := zap.NewAtomicLevel()

	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, nil, err
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

// init initiate MyLogger when this package is imported
func init() {
	MyLogger, MyProps, _ = newLogger()
	ReplaceGlobals(MyLogger, MyProps)
}
