package retrymiddleware

import (
	"context"
	"sync"

	"github.com/Fishwaldo/go-taskmanager"
	"github.com/Fishwaldo/go-taskmanager/joberrors"
	schedmetrics "github.com/Fishwaldo/go-taskmanager/metrics"
	"github.com/armon/go-metrics"
)

var _ taskmanager.RetryMiddleware = (*RetryCountLimit)(nil)

type retryCountCtxKey struct{}

// RetryConstantBackoff is a Middleware that will retry jobs after failures.
// By Default, it runs after Panics, Deferred Jobs (by other Middleware) or if OverLapped Jobs are prohibited.
// It uses a Constant Backoff Scheme and is implemented by github.com/cenkalti/backoff/v4
type RetryCountLimit struct {
	mx  sync.RWMutex
	max int
	RetryMiddlewareOptions
}

type retryCountCtx struct {
	attempts int
}

func (ebh *RetryCountLimit) getCtx(s *taskmanager.Task) (*retryCountCtx, bool) {
	bo, ok := s.Ctx.Value(retryCountCtxKey{}).(*retryCountCtx)
	return bo, ok
}

func (ebh *RetryCountLimit) setCtx(s *taskmanager.Task, bo *retryCountCtx) {
	s.Ctx = context.WithValue(s.Ctx, retryCountCtxKey{}, bo)
}

//Handler Run the Contant Backoff Handler
func (ebh *RetryCountLimit) Handler(s *taskmanager.Task, prerun bool, e error) (retry taskmanager.RetryResult, err error) {
	ebh.mx.Lock()
	defer ebh.mx.Unlock()
	bo, ok := ebh.getCtx(s)
	if !ok {
		s.Logger.Error("RetryCountLimit Not Reset/Initialzied")
		return taskmanager.RetryResult{Result: taskmanager.RetryResult_NextMW}, joberrors.FailedJobError{ErrorType: joberrors.Error_Middleware, Message: "RetryCountLimit Not Reset/Initialzied"}
	}
	if ebh.shouldHandleState(e) {
		bo.attempts++
		if bo.attempts > ebh.max {
			s.Logger.Warn("Exceeded Max Number of Attempts: %d (%d limit)", bo.attempts, ebh.max)
			metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_MW_RetryLimit_Hit), 1, []metrics.Label{{Name: "id", Value: s.GetID()}})
			return taskmanager.RetryResult{Result: taskmanager.RetryResult_NoRetry}, nil
		} else {
			s.Logger.Info("Retrying Job: Attempt %d - %d limit", bo.attempts, ebh.max)
			return taskmanager.RetryResult{Result: taskmanager.RetryResult_NextMW}, nil
		}
	}
	return taskmanager.RetryResult{Result: taskmanager.RetryResult_NextMW}, nil
}

func (ebh *RetryCountLimit) Reset(s *taskmanager.Task) (ok bool) {
	bo, ok := ebh.getCtx(s)
	if ok {
		bo.attempts = 0
	} else {
		ebh.setCtx(s, &retryCountCtx{attempts: 0})
	}
	return true
}

func (ebh *RetryCountLimit) Initilize(s *taskmanager.Task) {
	ebh.Reset(s)
}

//NewDefaultRetryConstantBackoff Create a Constant Backoff Handler with 1 second
func NewDefaultRetryCountLimit() *RetryCountLimit {
	val := NewRetryRetryCountLimit(10)
	return val
}

//NewConstandBackoffMW Create a Constant Backoff Handler with Specified Duration
func NewRetryRetryCountLimit(limit int) *RetryCountLimit {
	val := RetryCountLimit{
		max: limit,
	}
	val.handleDeferred = true
	val.handleOverlap = true
	val.handlePanic = true
	return &val
}
