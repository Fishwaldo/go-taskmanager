package taskmanager

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"os"

	"github.com/Fishwaldo/go-taskmanager/utils"
)

type Logger interface {
	Trace(message string, params ...interface{})
	Debug(message string, params ...interface{})
	Info(message string, params ...interface{})
	Warn(message string, params ...interface{})
	Error(message string, params ...interface{})
	Fatal(message string, params ...interface{})
	Panic(message string, params ...interface{})
	New(name string) (l Logger)
	With(key string, value interface{}) (l Logger)
	Sync()
}

//DefaultLogger uses Golang Standard Logging Libary
func DefaultLogger() (l *StdLogger) {
	stdlogger := log.New(os.Stderr, "sched  - ", log.LstdFlags)
	stdlog := &StdLogger{Log: *stdlogger, keys: make(map[string]interface{})}
	return stdlog
}

type StdLogger struct {
	Log   log.Logger
	keys  map[string]interface{}
	mx    sync.Mutex
	level Log_Level
}

type Log_Level int

const (
	LOG_TRACE Log_Level = iota
	LOG_DEBUG
	LOG_INFO
	LOG_WARN
	LOG_ERROR
	LOG_FATAL
	LOG_PANIC
)

func (l *StdLogger) Trace(message string, params ...interface{}) {
	if l.level <= LOG_TRACE {
		l.Log.Printf("TRACE: %s - %s", fmt.Sprintf(message, params...), l.getKeys())
	}
}
func (l *StdLogger) Debug(message string, params ...interface{}) {
	if l.level <= LOG_DEBUG {
		l.Log.Printf("DEBUG: %s - %s", fmt.Sprintf(message, params...), l.getKeys())
	}
}
func (l *StdLogger) Info(message string, params ...interface{}) {
	if l.level <= LOG_INFO {
		l.Log.Printf("INFO: %s - %s", fmt.Sprintf(message, params...), l.getKeys())
	}
}
func (l *StdLogger) Warn(message string, params ...interface{}) {
	if l.level <= LOG_WARN {
		l.Log.Printf("WARN: %s - %s", fmt.Sprintf(message, params...), l.getKeys())
	}
}
func (l *StdLogger) Error(message string, params ...interface{}) {
	if l.level <= LOG_ERROR {
		l.Log.Printf("ERROR: %s - %s", fmt.Sprintf(message, params...), l.getKeys())
	}
}
func (l *StdLogger) Fatal(message string, params ...interface{}) {
	l.Log.Fatal(fmt.Printf("FATAL: %s - %s", fmt.Sprintf(message, params...), l.getKeys()))
}
func (l *StdLogger) Panic(message string, params ...interface{}) {
	l.Log.Panic(fmt.Printf("PANIC: %s - %s", fmt.Sprintf(message, params...), l.getKeys()))
}
func (l *StdLogger) New(name string) Logger {
	//nl := &StdLogger{keys: l.keys}
	nl := &StdLogger{level: l.level}
	nl.Log.SetPrefix(fmt.Sprintf("%s.%s", l.Log.Prefix(), name))
	nl.Log.SetFlags(l.Log.Flags())
	nl.Log.SetOutput(l.Log.Writer())
	return nl
}
func (l *StdLogger) With(key string, value interface{}) Logger {
	l.mx.Lock()
	defer l.mx.Unlock()
	stdlog := &StdLogger{level: l.level, keys: utils.CopyableMap(l.keys).DeepCopy()}
	stdlog.Log.SetPrefix(l.Log.Prefix())
	stdlog.Log.SetFlags(l.Log.Flags())
	stdlog.Log.SetOutput(l.Log.Writer())
	stdlog.keys[key] = value
	return stdlog
}
func (l *StdLogger) Sync() {
	// do nothing
}

func (l *StdLogger) getKeys() (message string) {
	l.mx.Lock()
	defer l.mx.Unlock()
	msg, err := json.Marshal(l.keys)
	if err == nil {
		return string(msg)
	}
	return err.Error()
}

func (l *StdLogger) SetLevel(level Log_Level) {
	l.level = level
}

func (l *StdLogger) GetLevel() (level Log_Level) {
	return l.level
}
