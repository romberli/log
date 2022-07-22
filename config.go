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
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pingcap/errors"
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
	DefaultLogLevel            = "info"
	DefaultDisableTimestamp    = false
	DefaultDisableDoubleQuotes = false
	DefaultDisableEscape       = false
)

var (
	ErrEmptyLogFileName    = "Log file name could NOT be an empty string."
	ErrNotValidLogFileName = "Log file name must be either unix or windows path format, %s is not valid."
)

// FileLogConfig serializes file log related config in yaml/json.
type FileLogConfig struct {
	FileName             string
	MaxSize              int
	MaxDays              int
	MaxBackups           int
	BackupFileNameOption Option
}

// NewFileLogConfig creates a FileLogConfig.
func NewFileLogConfig(fileName string, maxSize, maxDays, maxBackups int, backupFileNameOption Option) (fileLogConfig *FileLogConfig, err error) {
	fileName = strings.TrimSpace(fileName)

	if fileName == "" {
		return nil, errors.New(fmt.Sprintf(ErrEmptyLogFileName))
	}

	valid, _ := govalidator.IsFilePath(fileName)
	if !valid {
		return nil, errors.New(fmt.Sprintf(ErrNotValidLogFileName, fileName))
	}

	fileLogConfig = &FileLogConfig{
		FileName:             fileName,
		MaxSize:              maxSize,
		MaxDays:              maxDays,
		MaxBackups:           maxBackups,
		BackupFileNameOption: backupFileNameOption,
	}

	return fileLogConfig, nil
}

// NewFileLogConfigWithDefaultFileName creates a FileLogConfig, if fileName is empty, it will use default file name.
func NewFileLogConfigWithDefaultFileName(fileName string, maxSize, maxDays, maxBackups int) (fileLogConfig *FileLogConfig, err error) {
	var baseDir string
	var logDir string

	fileName = strings.TrimSpace(fileName)

	if fileName == "" {
		baseDir, err = os.Getwd()
		if err != nil {
			return nil, errors.Trace(err)
		}

		logDir = path.Join(baseDir, "log")
		_, err = os.Stat(logDir)
		if err != nil {
			if os.IsNotExist(err) {
				_, err = os.Create(logDir)
				if err != nil {
					return nil, errors.Trace(err)
				}
			} else {
				return nil, errors.Trace(err)
			}
		}

		fileName = path.Join(logDir, DefaultLogFileName)
	} else {
		logDir = path.Dir(fileName)
	}

	fileLogConfig = &FileLogConfig{
		FileName:             fileName,
		MaxSize:              maxSize,
		MaxDays:              maxDays,
		MaxBackups:           maxBackups,
		BackupFileNameOption: nil,
	}

	return fileLogConfig, nil
}

// NewEmptyFileLogConfig returns an empty *FileLogConfig
func NewEmptyFileLogConfig() *FileLogConfig {
	return &FileLogConfig{}
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
	// DisableDoubleQuote disables adding double-quotes to log entry
	DisableDoubleQuotes bool
	// DisableEscape disables escaping special characters like \n,\r...
	DisableEscape bool
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
		Level:               level,
		Format:              format,
		DisableTimestamp:    DefaultDisableTimestamp,
		File:                fileCfg,
		DisableDoubleQuotes: DefaultDisableDoubleQuotes,
		DisableEscape:       DefaultDisableEscape,
	}
}

// NewConfigWithStdout returns a *Config with given level and format
func NewConfigWithStdout(level, format string) *Config {
	return &Config{
		Level:  level,
		Format: format,
	}
}

// NewConfigWithFileLog returns a *Config with file options
func NewConfigWithFileLog(fileName, level, format string, maxSize, maxDays, maxBackups int, backupFileNameOption Option) (*Config, error) {
	fileCfg, err := NewFileLogConfig(fileName, maxSize, maxDays, maxBackups, backupFileNameOption)
	if err != nil {
		return nil, err
	}

	return &Config{
		Level:               level,
		Format:              format,
		DisableTimestamp:    DefaultDisableTimestamp,
		File:                *fileCfg,
		DisableDoubleQuotes: DefaultDisableDoubleQuotes,
		DisableEscape:       DefaultDisableEscape,
	}, nil
}

// SetDisableDoubleQuotes disables wrapping log content with double quotes
func (cfg *Config) SetDisableDoubleQuotes(disableDoubleQuotes bool) {
	cfg.DisableDoubleQuotes = disableDoubleQuotes
}

// SetDisableEscape disables escaping special characters of log content like \n,\r...
func (cfg *Config) SetDisableEscape(disableEscape bool) {
	cfg.DisableEscape = disableEscape
}

// buildOptions returns []zap.Option with options of config
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
			return zapcore.NewSamplerWithOptions(core, time.Second, cfg.Sampling.Initial, cfg.Sampling.Thereafter)
		}))
	}

	return opts
}

// ZapProperties records some information about zap
type ZapProperties struct {
	Core   zapcore.Core
	Syncer zapcore.WriteSyncer
	Level  zap.AtomicLevel
}

// Clone returns a fresh new *ZapProperties with same options,
// note that it will use the same syncer
func (props *ZapProperties) Clone() *ZapProperties {
	core := props.Core.With([]zapcore.Field{})
	level := zap.NewAtomicLevelAt(props.Level.Level())

	return &ZapProperties{
		core,
		props.Syncer,
		level,
	}
}

// SetCore sets the core
func (props *ZapProperties) SetCore(core zapcore.Core) {
	props.Core = core
}

// WithCore returns a fresh new *ZapProperties with given core
func (props *ZapProperties) WithCore(core zapcore.Core) *ZapProperties {
	level := zap.NewAtomicLevelAt(props.Level.Level())

	return &ZapProperties{
		core,
		props.Syncer,
		level,
	}
}

// newZapTextEncoder returns zapcore.Encoder with given config
func newZapTextEncoder(cfg *Config) zapcore.Encoder {
	return NewTextEncoder(cfg)
}
