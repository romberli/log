package log

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pingcap/errors"
)

const (
	backupTimeFormat       = "2006-01-02T15-04-05.000"
	backupTimeMinuteFormat = "200601021504"
	compressSuffix         = ".gz"
	defaultMaxSize         = 100
)

// ensure we always implement io.WriteCloser
var _ io.WriteCloser = (*Writer)(nil)

type Option func(name string) string

func DefaultRotateOption(name string) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	filename = filename[:len(filename)-len(ext)]
	ext = filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	timestamp := time.Now().Format(backupTimeMinuteFormat)

	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}

// Writer is an io.WriteCloser that writes to the specified filename.
//
// Writer opens or creates the logfile on first Write.  If the file exists and
// is less than MaxSize megabytes, lumberjack will open and append to that file.
// If the file exists and its size is >= MaxSize megabytes, the file is renamed
// by putting the current time in a timestamp in the name immediately before the
// file's extension (or the end of the filename if there's no extension). A new
// log file is then created using original filename.
//
// Whenever a write would cause the current log file exceed MaxSize megabytes,
// the current file is closed, renamed, and a new log file created with the
// original name. Thus, the filename you give Writer is always the "current" log
// file.
//
// Backups use the log file name given to Writer, in the form
// `name-timestamp.ext` where name is the filename without the extension,
// timestamp is the time at which the log was rotated formatted with the
// time.Time format of `2006-01-02T15-04-05.000` and the extension is the
// original extension.  For example, if your Writer.Filename is
// `/var/log/foo/server.log`, a backup created at 6:30pm on Nov 11 2016 would
// use the filename `/var/log/foo/server-2016-11-04T18-30-00.000.log`
//
// Cleaning Up Old Log Files
//
// Whenever a new logfile gets created, old log files may be deleted.  The most
// recent files according to the encoded timestamp will be retained, up to a
// number equal to MaxBackups (or all of them if MaxBackups is 0).  Any files
// with an encoded timestamp older than MaxAge days are deleted, regardless of
// MaxBackups.  Note that the time encoded in the timestamp is the rotation
// time, which may differ from the last time that file was written to.
//
// If MaxBackups and MaxAge are both 0, no old log files will be deleted.
type Writer struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	Filename string `json:"filename" yaml:"filename"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" yaml:"maxsize"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"maxage" yaml:"maxage"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxbackups" yaml:"maxbackups"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool `json:"localtime" yaml:"localtime"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `json:"compress" yaml:"compress"`

	// BackupFileNameOption is an optional function that returns the backup file name
	BackupFileNameOption Option

	size int64
	file *os.File
	mu   sync.Mutex

	millCh    chan bool
	startMill sync.Once
}

var (
	// currentTime exists so it can be mocked out by tests.
	currentTime = time.Now

	// osStat exists so it can be mocked out by tests.
	osStat = os.Stat

	// megabyte is the conversion factor between MaxSize and bytes.  It is a
	// variable so tests can mock it out and not need to write megabytes of data
	// to disk.
	megabyte = 1024 * 1024
)

// Write implements io.Writer.  If a write would cause the log file to be larger
// than MaxSize, the file is closed, renamed to include a timestamp of the
// current time, and a new log file is created using the original log file name.
// If the length of the write is greater than MaxSize, an error is returned.
func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	writeLen := int64(len(p))
	if writeLen > w.max() {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum file size %d", writeLen, w.max(),
		)
	}

	if w.file == nil {
		if err = w.openExistingOrNew(len(p)); err != nil {
			return 0, err
		}
	}

	if w.size+writeLen > w.max() {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	w.size += int64(n)

	return n, err
}

// Close implements io.Closer, and closes the current logfile.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.close()
}

// close closes the file if it is open.
func (w *Writer) close() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

// Rotate causes Writer to close the existing log file and immediately create a
// new one.  This is a helper function for applications that want to initiate
// rotations outside the normal rotation rules, such as in response to
// SIGHUP.  After rotating, this initiates compression and removal of old log
// files according to the configuration.
func (w *Writer) Rotate() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.rotate()
}

// rotate closes the current file, moves it aside with a timestamp in the name,
// (if it exists), opens a new file with the original filename, and then runs
// post-rotation processing and removal.
func (w *Writer) rotate() error {
	if err := w.close(); err != nil {
		return err
	}
	if err := w.openNew(); err != nil {
		return err
	}
	w.mill()
	return nil
}

// openNew opens a new log file for writing, moving any old log file out of the
// way. This method assumes the file has already been closed.
func (w *Writer) openNew() error {
	err := os.MkdirAll(w.dir(), 0744)
	if err != nil {
		return fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	name := w.filename()
	mode := os.FileMode(0644)
	info, err := osStat(name)
	if err == nil {
		// Copy the mode off the old logfile.
		mode = info.Mode()
		// move the existing file
		newname := w.backupName(name, w.LocalTime)
		if err := os.Rename(name, newname); err != nil {
			return fmt.Errorf("can't rename log file: %s", err)
		}

		// this is a no-op anywhere but linux
		if err := chown(name, info); err != nil {
			return err
		}
	}

	// we use truncate here because this should only get called when we've moved
	// the file ourselves. if someone else creates the file in the meantime,
	// just wipe out the contents.
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	w.file = f
	w.size = 0
	return nil
}

// backupName creates a new filename from the given name, inserting a timestamp
// between the filename and the extension, using the local time if requested
// (otherwise UTC).
func (w *Writer) backupName(name string, local bool) string {
	if w.BackupFileNameOption != nil {
		return w.BackupFileNameOption(name)
	}

	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	t := currentTime()
	if !local {
		t = t.UTC()
	}

	timestamp := t.Format(backupTimeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}

// openExistingOrNew opens the logfile if it exists and if the current write
// would not put it over MaxSize.  If there is no such file or the write would
// put it over the MaxSize, a new file is created.
func (w *Writer) openExistingOrNew(writeLen int) error {
	w.mill()

	filename := w.filename()
	info, err := osStat(filename)
	if os.IsNotExist(err) {
		return w.openNew()
	}
	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	if info.Size()+int64(writeLen) >= w.max() {
		return w.rotate()
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return w.openNew()
	}
	w.file = file
	w.size = info.Size()
	return nil
}

// genFilename generates the name of the logfile from the current time.
func (w *Writer) filename() string {
	if w.Filename != "" {
		return w.Filename
	}
	name := filepath.Base(os.Args[0]) + "-lumberjack.log"
	return filepath.Join(os.TempDir(), name)
}

// millRunOnce performs compression and removal of stale log files.
// Log files are compressed if enabled via configuration and old log
// files are removed, keeping at most l.MaxBackups files, as long as
// none of them are older than MaxAge.
func (w *Writer) millRunOnce() error {
	if w.MaxBackups == 0 && w.MaxAge == 0 && !w.Compress {
		return nil
	}

	files, err := w.oldLogFiles()
	if err != nil {
		return err
	}

	var compress, remove []logInfo

	if w.MaxBackups > 0 && w.MaxBackups < len(files) {
		preserved := make(map[string]bool)
		var remaining []logInfo
		for _, f := range files {
			// Only count the uncompressed log file or the
			// compressed log file, not both.
			fn := f.Name()
			if strings.HasSuffix(fn, compressSuffix) {
				fn = fn[:len(fn)-len(compressSuffix)]
			}
			preserved[fn] = true

			if len(preserved) > w.MaxBackups {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}
	if w.MaxAge > 0 {
		diff := time.Duration(int64(24*time.Hour) * int64(w.MaxAge))
		cutoff := currentTime().Add(-1 * diff)

		var remaining []logInfo
		for _, f := range files {
			if f.timestamp.Before(cutoff) {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}

	if w.Compress {
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), compressSuffix) {
				compress = append(compress, f)
			}
		}
	}

	for _, f := range remove {
		errRemove := os.Remove(filepath.Join(w.dir(), f.Name()))
		if err == nil && errRemove != nil {
			err = errRemove
		}
	}
	for _, f := range compress {
		fn := filepath.Join(w.dir(), f.Name())
		errCompress := compressLogFile(fn, fn+compressSuffix)
		if err == nil && errCompress != nil {
			err = errCompress
		}
	}

	return err
}

// millRun runs in a goroutine to manage post-rotation compression and removal
// of old log files.
func (w *Writer) millRun() {
	for range w.millCh {
		// what am I going to do, log this?
		_ = w.millRunOnce()
	}
}

// mill performs post-rotation compression and removal of stale log files,
// starting the mill goroutine if necessary.
func (w *Writer) mill() {
	w.startMill.Do(func() {
		w.millCh = make(chan bool, 1)
		go w.millRun()
	})
	select {
	case w.millCh <- true:
	default:
	}
}

// oldLogFiles returns the list of backup log files stored in the same
// directory as the current log file, sorted by ModTime
func (w *Writer) oldLogFiles() ([]logInfo, error) {
	files, err := ioutil.ReadDir(w.dir())
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	var logFiles []logInfo

	prefix, ext := w.prefixAndExt()

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if t, err := w.timeFromName(f.Name(), prefix, ext); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
		if t, err := w.timeFromName(f.Name(), prefix, ext+compressSuffix); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
		// error parsing means that the suffix at the end was not generated
		// by lumberjack, and therefore it's not a backup file.
	}

	sort.Sort(byFormatTime(logFiles))

	return logFiles, nil
}

// timeFromName extracts the formatted time from the filename by stripping off
// the filenames prefix and extension. This prevents someone's filename from
// confusing time.parse.
func (w *Writer) timeFromName(filename, prefix, ext string) (time.Time, error) {
	if !strings.HasPrefix(filename, prefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}
	if !strings.HasSuffix(filename, ext) {
		return time.Time{}, errors.New("mismatched extension")
	}
	ts := filename[len(prefix) : len(filename)-len(ext)]
	return time.Parse(backupTimeFormat, ts)
}

// max returns the maximum size in bytes of log files before rolling.
func (w *Writer) max() int64 {
	if w.MaxSize == 0 {
		return int64(defaultMaxSize * megabyte)
	}
	return int64(w.MaxSize) * int64(megabyte)
}

// dir returns the directory for the current filename.
func (w *Writer) dir() string {
	return filepath.Dir(w.filename())
}

// prefixAndExt returns the filename part and extension part from the Writer's
// filename.
func (w *Writer) prefixAndExt() (prefix, ext string) {
	filename := filepath.Base(w.filename())
	ext = filepath.Ext(filename)
	prefix = filename[:len(filename)-len(ext)] + "-"
	return prefix, ext
}

// compressLogFile compresses the given log file, removing the
// uncompressed log file if successful.
func compressLogFile(src, dst string) (err error) {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer func() { _ = f.Close() }()

	fi, err := osStat(src)
	if err != nil {
		return fmt.Errorf("failed to stat log file: %v", err)
	}

	if err := chown(dst, fi); err != nil {
		return fmt.Errorf("failed to chown compressed log file: %v", err)
	}

	// If this file already exists, we presume it was created by
	// a previous attempt to compress the log file.
	gzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
	if err != nil {
		return fmt.Errorf("failed to open compressed log file: %v", err)
	}
	defer func() { _ = gzf.Close() }()

	gz := gzip.NewWriter(gzf)

	defer func() {
		if err != nil {
			_ = os.Remove(dst)
			err = fmt.Errorf("failed to compress log file: %v", err)
		}
	}()

	if _, err := io.Copy(gz, f); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	if err := gzf.Close(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil {
		return err
	}

	return nil
}

// logInfo is a convenience struct to return the filename and its embedded
// timestamp.
type logInfo struct {
	timestamp time.Time
	os.FileInfo
}

// byFormatTime sorts by newest time formatted in the name.
type byFormatTime []logInfo

func (b byFormatTime) Less(i, j int) bool {
	return b[i].timestamp.After(b[j].timestamp)
}

func (b byFormatTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byFormatTime) Len() int {
	return len(b)
}
