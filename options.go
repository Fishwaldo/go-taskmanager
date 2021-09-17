package taskmanager

import (

)

type taskoptions struct {
	logger              Logger
	executationmiddlewares []ExecutionMiddleWare
	retryMiddlewares	   []RetryMiddleware
}


func defaultTaskOptions() *taskoptions {
	logger := DefaultLogger()
	return &taskoptions{
		logger:       logger,
	}
}

func defaultSchedOptions() *taskoptions {
	logger := DefaultLogger()
	return &taskoptions {
		logger: 	logger,
	}
}

// Option to customize schedule behavior, check the sched.With*() functions that implement Option interface for the
// available options
type Option interface {
	apply(*taskoptions)
}

type loggerOption struct {
	Logger Logger
}

func (l loggerOption) apply(opts *taskoptions) {
	opts.logger = l.Logger
}

//WithLogger Use the supplied Logger as the logger.
func WithLogger(logger Logger) Option {
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

