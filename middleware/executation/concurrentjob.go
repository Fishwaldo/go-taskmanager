package executionmiddleware

import (
	"context"
	"sync"

	"github.com/Fishwaldo/go-taskmanager"
	"github.com/Fishwaldo/go-taskmanager/joberrors"
	schedmetrics "github.com/Fishwaldo/go-taskmanager/metrics"
	"github.com/armon/go-metrics"
)

var _ taskmanager.ExecutionMiddleWare = (*ConcurrentJobBlocker)(nil)

type hasCJBtxKey struct{}

// ConcurrentJobBlocker is a Middleware that will defer a job if one is already running
type ConcurrentJobBlocker struct {
	mx sync.Mutex
}

type cjllock struct {
	running bool
}

func (hth *ConcurrentJobBlocker) getTagCtx(s *taskmanager.Task) *cjllock {
	cjl, ok := s.Ctx.Value(hasCJBtxKey{}).(*cjllock)
	if !ok {
		return nil
	}
	return cjl
}

func (hth *ConcurrentJobBlocker) setTagCtx(s *taskmanager.Task, cjl *cjllock) {
	s.Ctx = context.WithValue(s.Ctx, hasCJBtxKey{}, cjl)
}

func (hth *ConcurrentJobBlocker) PreHandler(s *taskmanager.Task) (taskmanager.MWResult, error) {
	hth.mx.Lock()
	defer hth.mx.Unlock()
	cjl := hth.getTagCtx(s)
	if cjl != nil {
		s.Logger.V(1).Info("Concurrent Job Lock", "locked", cjl.running)
		if cjl.running {
			metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_MW_ConcurrentJob_Blocked), 1, []metrics.Label{{Name: "id", Value: s.GetID()}})
			return taskmanager.MWResult{Result: taskmanager.MWResult_Defer}, joberrors.FailedJobError{Message: "Job Already Running", ErrorType: joberrors.Error_ConcurrentJob}
		} else {
			cjl.running = true
		}
	} else {
		return taskmanager.MWResult{Result: taskmanager.MWResult_NextMW}, joberrors.FailedJobError{Message: "ConcurrentJobBlocker Not Initilzied", ErrorType: joberrors.Error_Middleware}
	}
	return taskmanager.MWResult{Result: taskmanager.MWResult_NextMW}, nil
}
func (hth *ConcurrentJobBlocker) PostHandler(s *taskmanager.Task, err error) taskmanager.MWResult {
	hth.mx.Lock()
	defer hth.mx.Unlock()
	cjl := hth.getTagCtx(s)
	if cjl != nil {
		s.Logger.V(1).Info("Concurrent Job Lock", "locked", cjl.running)
		if !cjl.running {
			s.Logger.Info("Job Was Not Locked")
			return taskmanager.MWResult{Result: taskmanager.MWResult_NextMW}
		} else {
			cjl.running = false
		}
	}
	return taskmanager.MWResult{Result: taskmanager.MWResult_NextMW}
}
func (hth *ConcurrentJobBlocker) Initilize(s *taskmanager.Task) {
	hth.mx.Lock()
	defer hth.mx.Unlock()
	hth.setTagCtx(s, &cjllock{running: false})
}

func (hth *ConcurrentJobBlocker) Reset(s *taskmanager.Task) {
	hth.mx.Lock()
	defer hth.mx.Unlock()
	cjl := hth.getTagCtx(s)
	cjl.running = false
}

// NewCJLock Create a new ConcurrentJob Lock Middleware
func NewCJLock() *ConcurrentJobBlocker {
	return &ConcurrentJobBlocker{}
}
