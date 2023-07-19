package log

import (
	"io"
	"os"

	"github.com/pingcap/errors"
	"github.com/romberli/go-multierror"
	"go.uber.org/zap/zapcore"
)

type WriteSyncer struct {
	io.Writer
	ws zapcore.WriteSyncer
}

// NewWriteSyncer returns a new zapcore.WriteSyncer
func NewWriteSyncer(w io.Writer) zapcore.WriteSyncer {
	return newWriteSyncer(w)
}

// NewStdoutWriteSyncer returns a zapcore.WriteSyncer using os.Stdout
func NewStdoutWriteSyncer() zapcore.WriteSyncer {
	return newWriteSyncer(os.Stdout)
}

// newWriteSyncer returns a new *WriterSyncer
func newWriteSyncer(w io.Writer) *WriteSyncer {
	return &WriteSyncer{
		Writer: w,
		ws:     zapcore.AddSync(w),
	}
}

// GetWriter returns the underlying writer of the WriteSyncer
func (ws WriteSyncer) GetWriter() io.Writer {
	return ws.Writer
}

// GetWriteSyncer returns the underlying WriteSyncer of the WriteSyncer
func (ws WriteSyncer) GetWriteSyncer() zapcore.WriteSyncer {
	return ws.ws
}

// Sync implements zapcore.WriteSyncer
func (ws WriteSyncer) Sync() error {
	return ws.ws.Sync()
}

type MultiWriteSyncer []zapcore.WriteSyncer

// NewMultiWriteSyncer creates a WriteSyncer that duplicates its writes
// and sync calls, much like io.MultiWriter.
func NewMultiWriteSyncer(ws ...zapcore.WriteSyncer) zapcore.WriteSyncer {
	if len(ws) == 1 {
		return ws[0]
	}
	// Copy to protect against https://github.com/golang/go/issues/7809
	return MultiWriteSyncer(append([]zapcore.WriteSyncer(nil), ws...))
}

// List lists write syncer of MultiWriteSyncer.
func (ws MultiWriteSyncer) List() []zapcore.WriteSyncer {
	var result []zapcore.WriteSyncer

	ws.list(&result)

	return result
}

// list lists write syncer of MultiWriteSyncer, it is the implementation of List(),
// if listed syncer is also a MultiWriteSyncer, it will call recursively.
func (ws MultiWriteSyncer) list(syncerList *[]zapcore.WriteSyncer) {
	for _, syncer := range ws {
		s, ok := syncer.(MultiWriteSyncer)
		if ok {
			s.list(syncerList)
		} else {
			*syncerList = append(*syncerList, s)
		}
	}
}

// See https://golang.org/src/io/multi.go
// When not all underlying syncers write the same number of bytes,
// the smallest number is returned even though Write() is called on
// all of them.
func (ws MultiWriteSyncer) Write(p []byte) (int, error) {
	var writeErr *multierror.Error
	nWritten := 0
	for _, w := range ws {
		n, err := w.Write(p)
		if err != nil {
			writeErr = multierror.Append(writeErr, errors.Trace(err))
		}
		if nWritten == 0 && n != 0 {
			nWritten = n
		} else if n < nWritten {
			nWritten = n
		}
	}
	return nWritten, writeErr.ErrorOrNil()
}

func (ws MultiWriteSyncer) Sync() error {
	var err *multierror.Error
	for _, w := range ws {
		err = multierror.Append(err, w.Sync())
	}
	return err.ErrorOrNil()
}
