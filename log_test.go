package taskmanager

import (
	"bytes"
	"os"
	"regexp"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	logger := DefaultLogger()
	logger.SetLevel(LOG_TRACE)
	if logger.level != LOG_TRACE {
		t.Errorf("Can't Set Logging Level")
	}
	if logger.GetLevel() != LOG_TRACE {
		t.Error("GetLevel Didn't return Correct Logging Level")
	}
}

func captureOutput(l *StdLogger, f func()) string {
    var buf bytes.Buffer
    l.Log.SetOutput(&buf)
    f()
    l.Log.SetOutput(os.Stderr)
    return buf.String()
}
func TestLogTrace(t *testing.T) {
	logger := DefaultLogger()
	logger.SetLevel(LOG_TRACE)
	output := captureOutput(logger, func() { 
		logger.Trace("Hello %s", "world")
	})
	validmsg := regexp.MustCompile(`^.* TRACE: Hello world \- {}`)
	if !validmsg.MatchString(output) {
		t.Errorf("Log Trace Failed: %s", output)
	}
}
func TestLogDebug(t *testing.T) {
	logger := DefaultLogger()
	logger.SetLevel(LOG_TRACE)
	output := captureOutput(logger, func() { 
		logger.Debug("Hello %s", "world")
	})
	validmsg := regexp.MustCompile(`^.* DEBUG: Hello world \- {}`)
	if !validmsg.MatchString(output) {
		t.Errorf("Log Debug Failed: %s", output)
	}
}

func TestLogInfo(t *testing.T) {
	logger := DefaultLogger()
	logger.SetLevel(LOG_TRACE)
	output := captureOutput(logger, func() { 
		logger.Info("Hello %s", "world")
	})
	validmsg := regexp.MustCompile(`^.* INFO: Hello world \- {}`)
	if !validmsg.MatchString(output) {
		t.Errorf("Log Info Failed: %s", output)
	}
}

func TestLogWarn(t *testing.T) {
	logger := DefaultLogger()
	logger.SetLevel(LOG_TRACE)
	output := captureOutput(logger, func() { 
		logger.Warn("Hello %s", "world")
	})
	validmsg := regexp.MustCompile(`^.* WARN: Hello world \- {}`)
	if !validmsg.MatchString(output) {
		t.Errorf("Log Warn Failed: %s", output)
	}
}

func TestLogError(t *testing.T) {
	logger := DefaultLogger()
	logger.SetLevel(LOG_TRACE)
	output := captureOutput(logger, func() { 
		logger.Error("Hello %s", "world")
	})
	validmsg := regexp.MustCompile(`^.* ERROR: Hello world \- {}`)
	if !validmsg.MatchString(output) {
		t.Errorf("Log Error Failed: %s", output)
	}
}

func TestLogFatal(t *testing.T) {
	//logger := DefaultLogger()
	//logger.SetLevel(LOG_TRACE)
	//output := captureOutput(logger, func() { 
	//	logger.Fatal("Hello %s", "world")
	//})
	//validmsg := regexp.MustCompile(`^.* FATAL: Hello world \- {}`)
	//if !validmsg.MatchString(output) {
	//	t.Errorf("Log Fatal Failed: %s", output)
	//}
}

func TestLogPanic(t *testing.T) {
	logger := DefaultLogger()
	logger.SetLevel(LOG_TRACE)
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("Log Panic Recovery Failed")
		}
	}()
	output := captureOutput(logger, func() { 
		logger.Panic("Hello %s", "world")
	})

	validmsg := regexp.MustCompile(`^.* PANIC: Hello world \- {}`)
	if !validmsg.MatchString(output) {
		t.Errorf("Log Panic Failed: %s", output)
	}
}
