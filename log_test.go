package log

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newRoutine(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	t.Log("new routine test")
	Debug("this is new routine debug message")
	Info("this is new routine info message")
	Warn("this is new routine warn message")
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
	asst.Nil(err, "init logger failed")
	t.Log("==========init logger completed==========\n")

	// print log
	t.Log("==========print main log entry started==========")
	Debug("this is main debug message")
	Info("this is main info message")
	MyLogger.Warnf("this is main warn message %s", "sss")
	MyLogger.Error("this is main error message")
	// MyLogger.Fatal("this is main fatal message")
	t.Log("==========print main log entry completed==========")

	t.Log("==========print goroutine log entry started==========")

	wg.Add(1)
	go func() {
		defer wg.Done()
		t.Log("goroutine test")
		Debug("this is goroutine debug message")
		MyLogger.Info("this is goroutine info message")
		MyLogger.Warn("this is goroutine warn message")
	}()

	wg.Add(1)
	go newRoutine(t, wg)
	wg.Wait()
	t.Log("==========print goroutine log entry completed==========")

	wg.Wait()
}
