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

import "go.uber.org/zap/zapcore"

// textIOCore is a copy of zapcore.ioCore that only accept *textEncoder
// it can be removed after https://github.com/uber-go/zap/pull/685 be merged
type textIOCore struct {
	zapcore.LevelEnabler
	enc *textEncoder
	out zapcore.WriteSyncer
}

// NewTextCore creates a Core that writes logs to a WriteSyncer.
func NewTextCore(enc *textEncoder, ws zapcore.WriteSyncer, enab zapcore.LevelEnabler) zapcore.Core {
	return &textIOCore{
		LevelEnabler: enab,
		enc:          enc,
		out:          ws,
	}
}

func (c *textIOCore) GetWriterSyncer() zapcore.WriteSyncer {
	return c.out
}

func (c *textIOCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	// it's different to ioCore, here call textEncoder#addFields to fix https://github.com/pingcap/log/issues/3
	clone.enc.addFields(fields)
	return clone
}

func (c *textIOCore) Syncer() zapcore.WriteSyncer {
	return c.out
}

func (c *textIOCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *textIOCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	_, err = c.out.Write(buf.Bytes())
	buf.Free()
	if err != nil {
		return err
	}
	if ent.Level > zapcore.ErrorLevel {
		// Since we may be crashing the program, sync the output. Ignore Sync
		// errors, pending a clean solution to issue https://github.com/uber-go/zap/issues/370.
		_ = c.Sync()
	}
	return nil
}

func (c *textIOCore) Sync() error {
	return c.out.Sync()
}

func (c *textIOCore) clone() *textIOCore {
	return &textIOCore{
		LevelEnabler: c.LevelEnabler,
		enc:          c.enc.Clone().(*textEncoder),
		out:          c.out,
	}
}

// SetTimeFormat sets the time format to the encoder
func (c *textIOCore) SetTimeFormat(timeFormat string) {
	c.enc.SetTimeFormat(timeFormat)
}

// SetSeperator sets the seperator to the encoder
func (c *textIOCore) SetSeperator(seperator string) {
	c.enc.SetSeperator(seperator)
}

// SetDisableDoubleQuotes disables wrapping log content with double quotes
func (c *textIOCore) SetDisableDoubleQuotes(disableDoubleQuotes bool) {
	c.enc.SetDisableDoubleQuotes(disableDoubleQuotes)
}

// SetDisableEscape disables escaping special characters of log content like \n,\r...
func (c *textIOCore) SetDisableEscape(disableEscape bool) {
	c.enc.SetDisableEscape(disableEscape)
}

func (c *textIOCore) ListWriteSyncer() []zapcore.WriteSyncer {
	multiWriteSyncer, ok := c.out.(MultiWriteSyncer)
	if ok {
		return multiWriteSyncer.List()
	}

	return []zapcore.WriteSyncer{c.out}
}

func (c *textIOCore) AddWriteSyncer(ws zapcore.WriteSyncer) {
	syncerList := c.ListWriteSyncer()
	syncerList = append(syncerList, ws)
	c.out = NewMultiWriteSyncer(syncerList...)
}
