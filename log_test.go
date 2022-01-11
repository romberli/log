package log

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	perr "github.com/pkg/errors"
	"go.uber.org/zap"

	// "go.uber.org/zap"

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
	fileName := "/tmp/run.log"
	maxSize := 1
	maxDays := 1
	maxBackups := 2

	// init logger
	t.Log("==========init logger started==========")
	asst.Nil(err, "init file log config failed")

	logConfig, err = NewConfigWithFileLog(fileName, level, format, maxSize, maxDays, maxBackups)
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
	return perr.WithStack(funcA())
}

func funcC() error {
	return perr.WithStack(funcB())
}

func TestLogStack(t *testing.T) {
	SetDisableEscape(true)
	SetDisableDoubleQuotes(true)
	err := funcC()
	// MyLogger = MyLogger.WithOptions(zap.AddCaller())
	// MyLogger.Error(fmt.Sprintf("ttt: %s", err.Error()))
	MyLogger.Error("ttt: ", zap.Error(err))
	// MyLogger.Errorf("", zap.Error(err).String)
	t.Log("=======")
	Error(err.Error())
}
