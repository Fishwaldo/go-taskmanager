package retrymiddleware

import (
	"context"
	"sync"

	"github.com/Fishwaldo/go-taskmanager"
	"github.com/Fishwaldo/go-taskmanager/joberrors"
	schedmetrics "github.com/Fishwaldo/go-taskmanager/metrics"
	"github.com/armon/go-metrics"
	"github.com/cenkalti/backoff/v4"
)

var _ taskmanager.RetryMiddleware = (*RetryExponentialBackoff)(nil)

type eboCtxKey struct{}

// RetryExponentialBackoff will retry a job using a exponential backoff
// By Default, it runs after Panics, Deferred Jobs (by other Middleware) or if OverLapped Jobs are prohibited.
// It uses a Exponential Backoff Scheme and is implemented by github.com/cenkalti/backoff/v4
type RetryExponentialBackoff struct {
	mx sync.RWMutex
	bo *backoff.ExponentialBackOff
	RetryMiddlewareOptions
}

func (ebh *RetryExponentialBackoff) getCtx(s *taskmanager.Task) (*backoff.ExponentialBackOff, bool) {
	bo, ok := s.Ctx.Value(eboCtxKey{}).(*backoff.ExponentialBackOff)
	return bo, ok
}

func (ebh *RetryExponentialBackoff) setCtx(s *taskmanager.Task, bo *backoff.ExponentialBackOff) {
	s.Ctx = context.WithValue(s.Ctx, eboCtxKey{}, bo)
}

//Handler Run the ExponentialBackoff
func (ebh *RetryExponentialBackoff) Handler(s *taskmanager.Task, prerun bool, e error) (retry taskmanager.RetryResult, err error) {
	ebh.mx.Lock()
	defer ebh.mx.Unlock()
	bo, ok := ebh.getCtx(s)
	if !ok {
		s.Logger.Error(nil, "RetryExponentialBackoff Not Reset/Initialzied")
		return taskmanager.RetryResult{Result: taskmanager.RetryResult_NextMW}, joberrors.FailedJobError{ErrorType: joberrors.Error_Middleware, Message: "RetryExponentialBackoff Not Reset/Initialzied"}
	}
	if ebh.shouldHandleState(e) {
		next := bo.NextBackOff()
		s.Logger.Info("Exponential BO Handler Retrying Job in %s", next)
		metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_MW_ExpBackoff_Retries), 1, []metrics.Label{{Name: "id", Value: s.GetID()}})
		return taskmanager.RetryResult{Result: taskmanager.RetryResult_Retry, Delay: next}, nil
	}
	return taskmanager.RetryResult{Result: taskmanager.RetryResult_NextMW}, nil
}
func (ebh *RetryExponentialBackoff) Reset(s *taskmanager.Task) (ok bool) {
	bo, ok := ebh.getCtx(s)
	if ok {
		bo.Reset()
	} else {
		bo := backoff.NewExponentialBackOff()
		bo.InitialInterval = ebh.bo.InitialInterval
		bo.MaxElapsedTime = ebh.bo.MaxElapsedTime
		bo.MaxInterval = ebh.bo.MaxInterval
		bo.Multiplier = ebh.bo.Multiplier
		bo.RandomizationFactor = ebh.bo.RandomizationFactor
		ebh.setCtx(s, bo)
	}
	return true
}

func (ebh *RetryExponentialBackoff) Initilize(s *taskmanager.Task) {
	ebh.Reset(s)
}

//NewDefaultExponentialBackoffMW Create a ExponentialBackoff Handler with Default settings
func NewDefaultRetryExponentialBackoff() *RetryExponentialBackoff {
	val := NewRetryExponentialBackoff(backoff.NewExponentialBackOff())
	return val
}

//NewExponentialBackoffMW Create a ExponentialBackoff Handler Custom Settings
//Uses github.com/cenkalti/backoff/v4 as the Backoff Implementation
func NewRetryExponentialBackoff(ebo *backoff.ExponentialBackOff) *RetryExponentialBackoff {
	val := RetryExponentialBackoff{
		bo: ebo,
	}
	val.handlePanic = true
	val.handleOverlap = true
	val.handleDeferred = true
	return &val
}
