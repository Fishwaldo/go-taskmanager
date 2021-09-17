package zaplog

import (
	"github.com/Fishwaldo/go-taskmanager"
	"go.uber.org/zap"
)

var _ taskmanager.Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	Zap *zap.SugaredLogger
}

func (l *ZapLogger) With(key string, value interface{}) taskmanager.Logger {
	nl := &ZapLogger{Zap: l.Zap.With(key, value)}
	return nl
}

func (l *ZapLogger) Trace(message string, params ...interface{}) {
	l.Zap.Debugf(message, params...)
}
func (l *ZapLogger) Debug(message string, params ...interface{}) {
	l.Zap.Debugf(message, params...)
}
func (l *ZapLogger) Info(message string, params ...interface{}) {
	l.Zap.Infof(message, params...)
}
func (l *ZapLogger) Warn(message string, params ...interface{}) {
	l.Zap.Warnf(message, params...)
}
func (l *ZapLogger) Error(message string, params ...interface{}) {
	l.Zap.Errorf(message, params...)
}
func (l *ZapLogger) Fatal(message string, params ...interface{}) {
	l.Zap.Fatalf(message, params...)
}
func (l *ZapLogger) Panic(message string, params ...interface{}) {
	l.Zap.Panicf(message, params...)
}
func (l *ZapLogger) New(name string) (nl taskmanager.Logger) {
	zl := &ZapLogger{Zap: l.Zap}
	zl.Zap.Named(name)
	return zl
}

func (l *ZapLogger) Sync() {

}

//DefaultLogger Return Default Sched Logger based on Zap's sugared logger
func NewZapLogger() *ZapLogger {
	// TODO control verbosity
	loggerBase, _ := zap.NewDevelopment(zap.AddCallerSkip(1), zap.AddStacktrace(zap.ErrorLevel))
	sugarLogger := loggerBase.Sugar()
	return &ZapLogger{
		Zap: sugarLogger,
	}
}
