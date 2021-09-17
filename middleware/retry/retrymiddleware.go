package retrymiddleware

import (
	"errors"

	"github.com/Fishwaldo/go-taskmanager/joberrors"
)

type RetryMiddlewareOptions struct {
	handlePanic    bool
	handleOverlap  bool
	handleDeferred bool
}

// HandlePanic Enable/Disable the ExponetialBackoff Handler for Panics
func (retryOptions *RetryMiddlewareOptions) HandlePanic(val bool) {
	retryOptions.handlePanic = val
}

// HandleOverlap Enable/Disable the ExponetialBackoff Handler for Overlapped Jobs
func (retryOptions *RetryMiddlewareOptions) HandleOverlap(val bool) {
	retryOptions.handleOverlap = val
}

// HandleDeferred Enable/Disable the ExponetialBackoff Handler for Deferred Jobs
func (retryOptions *RetryMiddlewareOptions) HandleDeferred(val bool) {
	retryOptions.handleDeferred = val
}

func (retryOptions *RetryMiddlewareOptions) shouldHandleState(e error) bool {
	var err joberrors.FailedJobError
	if errors.As(e, &err) {
		switch err.ErrorType {
		case joberrors.Error_Panic:
			if retryOptions.handlePanic {
				return true
			}
		case joberrors.Error_ConcurrentJob:
			if retryOptions.handleOverlap {
				return true
			}
		case joberrors.Error_DeferedJob:
			if retryOptions.handleDeferred {
				return true
			}
		}
	}
	return false
}
