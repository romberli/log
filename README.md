# log

a rotatable log

NOTE: use release v1.0.2 and above, do NOT use releases below v1.0.2, because the behavior of old and new releases are different, and there is no way to delete old releases, because sum.golang.org will cache the checksum of old releases permanently and can not be cleaned. 

## how to use
```
import github.com/romberli/log

func main() {
    level := "info"
    format := "text"
    fileName := "/tmp/run.log"
    maxSize := 100 // MB
    maxDays := 7
    maxBackups := 5
    
    logConfig, err := log.NewConfigWithFileLog(level, format, fileName, maxSize, maxDays, maxBackups)
    if err != nil {
        fmt.Printf("got error when creating log config.\n%s", err.Error())
    }
    
    _, _, err = log.InitLogger(logConfig)
    if err != nil {
        fmt.Printf("init logger failed.\n%s", err.Error())
    }
    
    log.Info("this is info message.")
    
    message := "some message"
    log.Warnf("this is warning message with variable message: %s", message")
}
```