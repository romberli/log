# log

log is a rotated log, when log file reaches ***maxSize***(default: 100MB), rotated files will be kept for ***maxDays***(default: 7 days) or ***maxBackups***(default: 5) whichever comes first.It wrapped zaplog and lumberjack.


**NOTE**: use release v1.0.2 and above, do **NOT** use releases below v1.0.2, because the behavior of old and new releases are different, and there is no way to delete old releases, because sum.golang.org will cache the checksum of old releases permanently and can not be cleaned. 

## how to use
```
import github.com/romberli/log

func main() {
    fileName := "/tmp/run.log"
    level := "info"
    format := "text"
    maxSize := 100 // MB
    maxDays := 7
    maxBackups := 5
    
    _, _, err = log.InitFileLogger(fileName, level, format, maxSize, maxDays, maxBackups)
    if err != nil {
        fmt.Printf("init logger failed.\n%s", err.Error())
    }
    
    log.Info("this is info message.")
    
    message := "some message"
    log.Warnf("this is warning message with variable message: %s", message")
}
```
or just specify log file name and use default value for other arguments
```
import github.com/romberli/log

func main() {
    fileName := "/tmp/run.log"

    _, _, err = log.InitFileLoggerWithDefault(fileName)
    if err != nil {
        fmt.Printf("init logger failed.\n%s", err.Error())
    }
    
    log.Info("this is info message.")
    
    message := "some message"
    log.Warnf("this is warning message with variable message: %s", message")
}
```
or specify nothing and log to stdout
```
import github.com/romberli/log

func main() {
    log.Info("this is info message.")
}
```
and you'll see the message in the console


## customize
if you don't want double-quotes in message content, you can disable it.
```
log.SetDisableDoubleQuotes(true)
```
if you don't want to escape some special characters(\n,\r,\t) in the message content, you can disable it.
```
log.SetDisableEscape(true)
```
if you want to add another output to current logger, for example: add stdout to current file logger
```
log.AddWriteSyncer(log.NewWriteSyncer(os.stdout))
```
alternatively
```
log.AddWriteSyncer(log.NewStdoutWriteSyncer())
```
this impacts globally, so if you want to add new output temporally, you can clone a temporal logger and add new output to the new logger.
```
log.Clone().AddWriteSyncer(log.NewStdoutWriteSyncer()).Info("this is cloned logger message")
```