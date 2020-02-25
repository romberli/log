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
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var MyLogger *zap.Logger
var MyProps *ZapProperties

func stringToLogLevel(level string) log.Level {
	switch strings.ToLower(level) {
	case "fatal":
		return log.FatalLevel
	case "error":
		return log.ErrorLevel
	case "warn", "warning":
		return log.WarnLevel
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	}

	return defaultLogLevel
}

// textFormatter is for compatibility with ngaut/log
type textFormatter struct {
	DisableTimestamp bool
	EnableEntryOrder bool
}

// Format implements logrus.Formatter
func (f *textFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	if !f.DisableTimestamp {
		fmt.Fprintf(b, "%s ", entry.Time.Format(defaultLogTimeFormat))
	}
	if file, ok := entry.Data["file"]; ok {
		fmt.Fprintf(b, "%s:%v:", file, entry.Data["line"])
	}
	fmt.Fprintf(b, " [%s] %s", entry.Level.String(), entry.Message)

	if f.EnableEntryOrder {
		keys := make([]string, 0, len(entry.Data))
		for k := range entry.Data {
			if k != "file" && k != "line" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(b, " %v=%v", k, entry.Data[k])
		}
	} else {
		for k, v := range entry.Data {
			if k != "file" && k != "line" {
				fmt.Fprintf(b, " %v=%v", k, v)
			}
		}
	}

	b.WriteByte('\n')

	return b.Bytes(), nil
}

func stringToLogFormatter(format string, disableTimestamp bool) log.Formatter {
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

	// use lumberjack to logrotate
	return &lumberjack.Logger{
		Filename:   cfg.FileName,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxDays,
		LocalTime:  true,
	}, nil
}

// newLogger returns a logger
func newLogger() (*zap.Logger, *ZapProperties, error) {
	var (
		err    error
		cfg    *Config
		stdOut zapcore.WriteSyncer
		close  func()
	)

	cfg = &Config{Level: "info", File: FileLogConfig{}}

	if stdOut, close, err = zap.Open([]string{"stdout"}...); err != nil {
		close()
		return nil, nil, errors.Trace(err)
	}

	if MyLogger, MyProps, err = InitLoggerWithWriteSyncer(
		cfg, stdOut, zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCaller(),
		zap.Development(),
	); err != nil {
		return nil, nil, errors.Trace(err)
	}

	return MyLogger, MyProps, err
}

// InitLogger initializes a zap logger.
func InitLogger(cfg *Config) (*zap.Logger, *ZapProperties, error) {
	var (
		err    error
		lg     *lumberjack.Logger
		stdOut zapcore.WriteSyncer
		close  func()
		output zapcore.WriteSyncer
	)

	if len(cfg.File.FileName) > 0 {
		if lg, err = initFileLog(&cfg.File); err != nil {
			return nil, nil, errors.Trace(err)
		}

		output = zapcore.AddSync(lg)
	} else {
		if stdOut, close, err = zap.Open([]string{"stdout"}...); err != nil {
			close()
			return nil, nil, errors.Trace(err)
		}
		output = stdOut
	}

	if MyLogger, MyProps, err = InitLoggerWithWriteSyncer(
		cfg, output, zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCaller(),
		zap.Development(),
	); err != nil {
		return nil, nil, errors.Trace(err)
	}

	ReplaceGlobals(MyLogger, MyProps)

	return MyLogger, MyProps, nil
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

// L returns the global Logger, which can be reconfigured with ReplaceGlobals.
// It's safe for concurrent use.
func L() *zap.Logger {
	return _globalL
}

// S returns the global SugaredLogger, which can be reconfigured with
// ReplaceGlobals. It's safe for concurrent use.
func S() *zap.SugaredLogger {
	return _globalS
}

func ReplaceGlobals(logger *zap.Logger, props *ZapProperties) {
	_globalL = logger
	_globalS = logger.Sugar()
	_globalP = props
}

var (
	_globalL, _globalP, _ = newLogger()
	_globalS              = _globalL.Sugar()
)

// Sync flushes any buffered log entries.
func Sync() error {
	err := L().Sync()
	if err != nil {
		return err
	}
	return S().Sync()
}
