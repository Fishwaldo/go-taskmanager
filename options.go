package taskmanager

import (
	"log"
	"os"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

type taskoptions struct {
	logger              logr.Logger
	executationmiddlewares []ExecutionMiddleWare
	retryMiddlewares	   []RetryMiddleware
}


func defaultTaskOptions() *taskoptions {
	logsink := log.New(os.Stdout, "", 0);
	return &taskoptions{
		logger:       stdr.New(logsink),
	}
}

func defaultSchedOptions() *taskoptions {
	logsink := log.New(os.Stdout, "", 0);
	return &taskoptions {
		logger: 	stdr.New(logsink),
	}
}

// Option to customize schedule behavior, check the sched.With*() functions that implement Option interface for the
// available options
type Option interface {
	apply(*taskoptions)
}

type loggerOption struct {
	Logger logr.Logger
}

func (l loggerOption) apply(opts *taskoptions) {
	opts.logger = l.Logger
}

//WithLogger Use the supplied Logger as the logger.
func WithLogger(logger logr.Logger) Option {
	return loggerOption{Logger: logger}
}

type executationMiddleWare struct {
	middleware ExecutionMiddleWare
}

func (l executationMiddleWare) apply(opts *taskoptions) {
	opts.executationmiddlewares = append(opts.executationmiddlewares, l.middleware)
}

func WithExecutationMiddleWare(handler ExecutionMiddleWare) Option {
	return executationMiddleWare{middleware: handler}
}

type retryMiddleware struct {
	middleware RetryMiddleware
}

func (l retryMiddleware) apply(opts *taskoptions) {
	opts.retryMiddlewares = append(opts.retryMiddlewares, l.middleware)
}

func WithRetryMiddleWare(handler RetryMiddleware) Option {
	return retryMiddleware{middleware: handler}
}

