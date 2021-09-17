package logruslog

import (
	"github.com/Fishwaldo/go-taskmanager"
	"github.com/sirupsen/logrus"
)

var _ taskmanager.Logger = (*LruLogger)(nil)

type LruLogger struct {
	Lru *logrus.Entry
}

func (l LruLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.Lru.Debugf(msg, keysAndValues...)
}
func (l LruLogger) Error(msg string, keysAndValues ...interface{}) {
	l.Lru.Errorf(msg, keysAndValues...)
}
func (l LruLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.Lru.Fatalf(msg, keysAndValues...)
}
func (l LruLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Lru.Infof(msg, keysAndValues...)
}
func (l LruLogger) Panic(msg string, keysAndValues ...interface{}) {
	l.Lru.Panicf(msg, keysAndValues...)
}
func (l LruLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.Lru.Warnf(msg, keysAndValues...)
}
func (l *LruLogger) With(key string, value interface{}) taskmanager.Logger {
	nl := &LruLogger{Lru: logrus.NewEntry(l.Lru.Logger)}
	nl.Lru = l.Lru.WithField(key, value)
	return nl
}
func (l *LruLogger) Trace(key string, args ...interface{}) {
	l.Lru.Tracef(key, args...)
}
func (l *LruLogger) Sync() {
}
func (l *LruLogger) New(name string) taskmanager.Logger {
	nl := &LruLogger{Lru: logrus.NewEntry(l.Lru.Logger)}
	return nl
}

//LogrusDefaultLogger Return Logger based on logrus with new instance
func LogrusDefaultLogger() taskmanager.Logger {
	// TODO control verbosity
	l := &LruLogger{Lru: logrus.NewEntry(logrus.New())}
	return l
}
