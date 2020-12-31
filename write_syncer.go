package log

import (
	"io"
	"os"

	"go.uber.org/multierr"
	"go.uber.org/zap/zapcore"
)

// NewWriteSyncer converts io.Writer to zapcore.WriteSyncer
func NewWriteSyncer(w io.Writer) zapcore.WriteSyncer {
	return zapcore.AddSync(w)
}

// NewStdoutWriteSyncer returns a zapcore.WriteSyncer using os.Stdout
func NewStdoutWriteSyncer() zapcore.WriteSyncer {
	return NewWriteSyncer(os.Stdout)
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

	ws.list(result)

	return result
}

// list lists write syncer of MultiWriteSyncer, it is the implementation of List(),
// if listed syncer is also a MultiWriteSyncer, it will call recursively.
func (ws MultiWriteSyncer) list(syncerList []zapcore.WriteSyncer) {
	for _, syncer := range ws {
		s, ok := syncer.(MultiWriteSyncer)
		if ok {
			s.list(syncerList)
		} else {
			syncerList = append(syncerList, s)
		}
	}
}

// See https://golang.org/src/io/multi.go
// When not all underlying syncers write the same number of bytes,
// the smallest number is returned even though Write() is called on
// all of them.
func (ws MultiWriteSyncer) Write(p []byte) (int, error) {
	var writeErr error
	nWritten := 0
	for _, w := range ws {
		n, err := w.Write(p)
		writeErr = multierr.Append(writeErr, err)
		if nWritten == 0 && n != 0 {
			nWritten = n
		} else if n < nWritten {
			nWritten = n
		}
	}
	return nWritten, writeErr
}

func (ws MultiWriteSyncer) Sync() error {
	var err error
	for _, w := range ws {
		err = multierr.Append(err, w.Sync())
	}
	return err
}
