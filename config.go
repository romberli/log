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
	"os"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// default log file name
	DefaultLogFileName = "run.log"
	// default maximum log file size
	DefaultLogMaxSize = 100 // MB
	// default maximum log file backup days
	DefaultLogMaxDays = 7
	// default maximum log file backup number
	DefaultLogMaxBackups = 5
	// default log format in string
	DefaultLogFormat = "text"
	// default log level in string
	DefaultLogLevel         = "info"
	DefaultDisableTimestamp = false
)

// FileLogConfig serializes file log related config in yaml/json.
type FileLogConfig struct {
	// Log filename.
	FileName string `yaml:"file-name" json:"file-name"`
	// Max size in MB for a log file.
	MaxSize int `yaml:"max-size" json:"max-size"`
	// Max log keep days.
	MaxDays int `yaml:"max-days" json:"max-days"`
	// Maximum number of old log files to retain.
	MaxBackups int `yaml:"max-backups" json:"max-backups"`
}

// NewFileLogConfig creates a FileLogConfig.
func NewFileLogConfig(fileName string, maxSize, maxDays, maxBackups int) (fileLogConfig *FileLogConfig, err error) {
	var baseDir string
	var logDir string

	fileName = strings.TrimSpace(fileName)

	if fileName == "" {
		if baseDir, err = os.Getwd(); err != nil {
			return nil, err
		}

		logDir = path.Join(baseDir, "log")

		if _, err := os.Stat(logDir); err != nil {
			if os.IsNotExist(err) {
				if _, err = os.Create(logDir); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		fileName = path.Join(logDir, DefaultLogFileName)
	} else {
		logDir = path.Dir(fileName)
	}

	fileLogConfig = &FileLogConfig{
		FileName:   fileName,
		MaxSize:    maxSize,
		MaxDays:    maxDays,
		MaxBackups: maxBackups,
	}

	return fileLogConfig, nil
}

// Config serializes log related config in yaml/json.
type Config struct {
	// Log level.
	Level string `yaml:"level" json:"level"`
	// Log format. one of json, text, or console.
	Format string `yaml:"format" json:"format"`
	// Disable automatic timestamps in output.
	DisableTimestamp bool `yaml:"disable-timestamp" json:"disable-timestamp"`
	// File log config.
	File FileLogConfig `yaml:"file" json:"file"`
	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `yaml:"development" json:"development"`
	// DisableCaller stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool `yaml:"disable-caller" json:"disable-caller"`
	// DisableStacktrace completely disables automatic stacktrace capturing. By
	// default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	DisableStacktrace bool `yaml:"disable-stacktrace" json:"disable-stacktrace"`
	// DisableErrorVerbose stops annotating logs with the full verbose error
	// message.
	DisableErrorVerbose bool `yaml:"disable-error-verbose" json:"disable-error-verbose"`
	// SamplingConfig sets a sampling strategy for the logger. Sampling caps the
	// global CPU and I/O load that logging puts on your process while attempting
	// to preserve a representative subset of your logs.
	//
	// Values configured here are per-second. See zapcore.NewSampler for details.
	Sampling *zap.SamplingConfig `yaml:"sampling" json:"sampling"`
}

// NewConfig creates a Config.
func NewConfig(level, format string, fileCfg FileLogConfig) *Config {
	return &Config{
		Level:            level,
		Format:           format,
		DisableTimestamp: DefaultDisableTimestamp,
		File:             fileCfg,
	}
}

// NewConfig creates a Config with file.
func NewConfigWithFileLog(fileName, level, format string, maxSize, maxDays, maxBackups int) (*Config, error) {
	fileCfg, err := NewFileLogConfig(fileName, maxSize, maxDays, maxBackups)
	if err != nil {
		return nil, err
	}

	return &Config{
		Level:            level,
		Format:           format,
		DisableTimestamp: DefaultDisableTimestamp,
		File:             *fileCfg,
	}, nil
}

// ZapProperties records some information about zap.
type ZapProperties struct {
	Core   zapcore.Core
	Syncer zapcore.WriteSyncer
	Level  zap.AtomicLevel
}

func newZapTextEncoder(cfg *Config) zapcore.Encoder {
	return NewTextEncoder(cfg)
}

func (cfg *Config) buildOptions(errSink zapcore.WriteSyncer) []zap.Option {
	opts := []zap.Option{zap.ErrorOutput(errSink)}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	stackLevel := zap.ErrorLevel
	if cfg.Development {
		stackLevel = zap.WarnLevel
	}
	if !cfg.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	if cfg.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, int(cfg.Sampling.Initial), int(cfg.Sampling.Thereafter))
		}))
	}

	return opts
}
