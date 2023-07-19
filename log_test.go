package log

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/pingcap/errors"
	"github.com/romberli/go-multierror"
	"github.com/stretchr/testify/assert"
)

func newRoutine(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	t.Log("new routine test")
	Debug("this is new routine debug message")
	Info("this is new routine info message")
	Warn("this is new routine warn message")
	Errorf("this is new routine error message.\n%s", "errorf")
}

func TestLog(t *testing.T) {
	var (
		err       error
		logConfig *Config
	)

	wg := &sync.WaitGroup{}
	asst := assert.New(t)

	level := "info"
	format := "text"
	fileName := "/tmp/run.log.crn"
	maxSize := 1
	maxDays := 1
	maxBackups := 2

	// init logger
	t.Log("==========init logger started==========")
	asst.Nil(err, "init file log config failed")

	logConfig, err = NewConfigWithFileLog(fileName, level, format, maxSize, maxDays, maxBackups, DefaultRotateOption)
	if err != nil {
		fmt.Printf("got error when creating log config.\n%s", err.Error())
	}

	MyLogger, MyProps, err = InitLoggerWithConfig(logConfig)
	SetDisableEscape(true)

	asst.Nil(err, "init logger failed")
	t.Log("==========init logger completed==========\n")

	// print log
	t.Log("==========print main log entry started==========")
	Debug("this is main debug message")
	Info("this is main \ninfo message")
	Infof("this is main info message %s", "infof")
	Errorf("this main error message %s", "errorf")
	MyLogger.Warn("this is mylogger main warn message")
	MyLogger.Warnf("this is mylogger main warn message %s", "warnf")
	MyLogger.Error("this is main mylogger error message")
	MyLogger.Errorf("this is mylogger main error message %s", "errorf")
	// MyLogger.Fatal("this is main mylogger fatal message")
	// MyLogger.Fatalf("this is mylogger main fatal message %s", "fatalf")
	t.Log("==========print main log entry completed==========")

	t.Log("==========print goroutine log entry started==========")

	SetDisableDoubleQuotes(true)
	// SetDisableEscape(true)
	wg.Add(1)
	go func() {
		defer wg.Done()
		t.Log("goroutine test")
		Debug("this is goroutine debug message")
		// MyLogger.SetDisableDoubleQuotes(false)
		MyLogger.Info("this is goroutine mylogger info message")
		MyLogger.Warn("this is goroutine mylogger warn message")
		Info("this is goroutine info message")
	}()

	wg.Add(1)
	go newRoutine(t, wg)
	wg.Wait()
	t.Log("==========print goroutine log entry completed==========")

	wg.Wait()

	t.Log("==========test clone==========")
	CloneStdoutLogger().Info("this is cloned logger info message")
	CloneStdoutLogger().Infof("this is cloned logger infof message")
	Info("this is original logger info message, which should not be printed to console")
	t.Log("==========test clone==========")

	t.Log("==========add stdout to logger started==========")
	MyLogger.CloneAndAddWriteSyncer(NewStdoutWriteSyncer()).Info("test CloneAndAddWriteSyncer()")
	MyLogger.Info("mylogger info message after test CloneAndAddWriteSyncer, this should not be printed to console")
	stdoutSyncer := NewStdoutWriteSyncer()
	AddWriteSyncer(stdoutSyncer)
	Info("this is main info message which prints to stdout")
	Error("this is main error message which prints to stdout")
	MyLogger.Error("this is main mylogger error message which prints to stdout")
	t.Log("==========add stdout to logger completed==========")
}

func funcA() error {
	return errors.New("function error")
}

func funcB() error {
	return errors.Trace(funcA())
}

func funcC() error {
	return errors.Trace(funcB())
}

type T struct {
	F func() error
}

func JSONMarshal() error {
	t := &T{funcA}
	_, err := json.Marshal(t)

	return errors.WithStack(err)
}

type ErrMessage struct {
	Raw string
}

func (em *ErrMessage) Error() string {
	return em.Raw
}

func TestLogStack(t *testing.T) {
	SetDisableEscape(true)
	SetDisableDoubleQuotes(true)
	err := funcC()
	merr := &multierror.Error{}
	merr = multierror.Append(merr, err)

	// MyLogger = MyLogger.WithOptions(zap.AddCaller())
	// MyLogger.Error(fmt.Sprintf("ttt: %s", err.Error()))
	// MyLogger.Error("func: ", zap.Error(err))
	// err = JSONMarshal()
	// msg := fmt.Sprintf("got error. err:\n%+v", err)
	// t.Log(msg)
	em := &ErrMessage{fmt.Sprintf("%+v", err)}
	MyLogger.Errorf("json: %s", em.Error())
	//
	// MyLogger.Errorf("json: %+v", merr.WrappedErrors())
	// t.Log("=======")
	//
	// err = fmt.Errorf("wrapped error. error:\n%+v", err)
	// MyLogger.Errorf("json: %+v", err)
	// Error(err.Error())
}

func TestLogRotate(t *testing.T) {
	asst := assert.New(t)

	level := "info"
	format := "text"
	fileName := "/tmp/run.log.current"
	maxSize := 1
	maxDays := 1
	maxBackups := 2

	logConfig, err := NewConfigWithFileLog(fileName, level, format, maxSize, maxDays, maxBackups, DefaultRotateOption)
	if err != nil {
		fmt.Printf("got error when creating log config.\n%s", err.Error())
	}

	MyLogger, MyProps, err = InitLoggerWithConfig(logConfig)
	SetDisableEscape(true)

	MyLogger.Info("before rotate")
	err = MyLogger.Rotate()
	asst.Nil(err, "rotate failed")
	MyLogger.Info("after rotate")
}

func TestGlobalLogger(t *testing.T) {
	level := "info"
	format := "text"
	fileName := "/tmp/run.log.current"
	maxSize := 1
	maxDays := 1
	maxBackups := 2

	logConfig, err := NewConfigWithFileLog(fileName, level, format, maxSize, maxDays, maxBackups, DefaultRotateOption)
	if err != nil {
		fmt.Printf("got error when creating log config.\n%s", err.Error())
	}
	MyLogger, MyProps, err = InitLoggerWithConfig(logConfig)
	SetDisableEscape(true)
	SetDisableDoubleQuotes(true)
	ReplaceGlobals(MyLogger, MyProps)

	Info("========")
	Info("info message")
	Infof("infof message")

	AddWriteSyncer(NewStdoutWriteSyncer())

	Info("info message after add stdout")
	Infof("infof message after add stdout")
}
