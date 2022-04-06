package retrymiddleware

import (
	"context"
	"sync"
	"time"

	"github.com/Fishwaldo/go-taskmanager"
	"github.com/Fishwaldo/go-taskmanager/joberrors"
	schedmetrics "github.com/Fishwaldo/go-taskmanager/metrics"
	"github.com/armon/go-metrics"
	"github.com/cenkalti/backoff/v4"
)

var _ taskmanager.RetryMiddleware = (*RetryConstantBackoff)(nil)

type constantCtxKey struct{}

// RetryConstantBackoff is a Middleware that will retry jobs after failures.
// By Default, it runs after Panics, Deferred Jobs (by other Middleware) or if OverLapped Jobs are prohibited.
// It uses a Constant Backoff Scheme and is implemented by github.com/cenkalti/backoff/v4
type RetryConstantBackoff struct {
	mx       sync.RWMutex
	interval time.Duration
	RetryMiddlewareOptions
}

func (ebh *RetryConstantBackoff) getCtx(s *taskmanager.Task) (*backoff.ConstantBackOff, bool) {
	bo, ok := s.Ctx.Value(constantCtxKey{}).(*backoff.ConstantBackOff)
	return bo, ok
}

func (ebh *RetryConstantBackoff) setCtx(s *taskmanager.Task, bo *backoff.ConstantBackOff) {
	s.Ctx = context.WithValue(s.Ctx, constantCtxKey{}, bo)
}

//Handler Run the Contant Backoff Handler
func (ebh *RetryConstantBackoff) Handler(s *taskmanager.Task, prerun bool, e error) (retry taskmanager.RetryResult, err error) {
	ebh.mx.Lock()
	defer ebh.mx.Unlock()
	bo, ok := ebh.getCtx(s)
	if !ok {
		s.Logger.Error(nil, "RetryConstantBackoff Not Reset/Initialzied")
		return taskmanager.RetryResult{Result: taskmanager.RetryResult_NextMW}, joberrors.FailedJobError{ErrorType: joberrors.Error_Middleware, Message: "RetryConstantBackoff Not Reset/Initialzied"}
	}
	if ebh.shouldHandleState(e) {
		next := bo.NextBackOff()
		s.Logger.Info("Constant BO Handler Retrying Job in %s", next)
		metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_MW_ConstantBackoff_Retries), 1, []metrics.Label{{Name: "id", Value: s.GetID()}})
		return taskmanager.RetryResult{Result: taskmanager.RetryResult_Retry, Delay: next}, nil
	}
	return taskmanager.RetryResult{Result: taskmanager.RetryResult_NextMW}, nil
}

func (ebh *RetryConstantBackoff) Reset(s *taskmanager.Task) (ok bool) {
	bo, ok := ebh.getCtx(s)
	if ok {
		bo.Reset()
	} else {
		bo := backoff.NewConstantBackOff(ebh.interval)
		ebh.setCtx(s, bo)
	}
	return true
}

func (ebh *RetryConstantBackoff) Initilize(s *taskmanager.Task) {
	ebh.Reset(s)
}

//NewDefaultRetryConstantBackoff Create a Constant Backoff Handler with 1 second
func NewDefaultRetryConstantBackoff() *RetryConstantBackoff {
	val := NewRetryConstantBackoff(1 * time.Second)
	return val
}

//NewConstandBackoffMW Create a Constant Backoff Handler with Specified Duration
func NewRetryConstantBackoff(interval time.Duration) *RetryConstantBackoff {
	val := RetryConstantBackoff{
		interval: interval,
	}
	val.handleDeferred = true
	val.handleOverlap = true
	val.handlePanic = true
	return &val
}
