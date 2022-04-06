package taskmanager

import (
	//	"errors"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/armon/go-metrics"
	"github.com/sasha-s/go-deadlock"
	"github.com/go-logr/logr"
	"github.com/Fishwaldo/go-taskmanager/job"
	"github.com/Fishwaldo/go-taskmanager/joberrors"
	schedmetrics "github.com/Fishwaldo/go-taskmanager/metrics"
)

// PreExecutionMiddleWare Interface for developing new Middleware
// Pre Executation Middleware is run before executing a job.
type ExecutionMiddleWare interface {
	PreHandler(s *Task) (MWResult, error)
	PostHandler(s *Task, err error) MWResult
	Reset(s *Task)
	Initilize(s *Task)
}

type RetryMiddleware interface {
	Handler(s *Task, prerun bool, e error) (retry RetryResult, err error)
	Reset(s *Task) (ok bool)
	Initilize(s *Task)
}

type RetryResult_Op int

const (
	RetryResult_Retry RetryResult_Op = iota
	RetryResult_NoRetry
	RetryResult_NextMW
)

type RetryResult struct {
	Result RetryResult_Op
	Delay  time.Duration
}

type MWResult_Op int

const (
	MWResult_Cancel MWResult_Op = iota
	MWResult_Defer
	MWResult_NextMW
)

type MWResult struct {
	Result MWResult_Op
}

type nextRuni struct {
	mx   deadlock.RWMutex
	time time.Time
}

func (nr *nextRuni) Get() time.Time {
	nr.mx.RLock()
	defer nr.mx.RUnlock()
	return nr.time
}

func (nr *nextRuni) Set(t time.Time) {
	nr.mx.Lock()
	defer nr.mx.Unlock()
	nr.time = t
}

// passing context to a function with variables: https://play.golang.org/p/SW7uoU_KjlR

// Task A Task is an object that wraps a Job (func(){}) and runs it on a schedule according to the supplied
// Timer; With the the ability to expose metrics, and write logs to indicate job health, state, and stats.
type Task struct {
	id string

	// Source function used to create job.Job
	jobSrcFunc func(ctx context.Context)

	// Timer used to trigger Jobs
	timer Timer

	// Next Scheduled Run
	nextRun nextRuni

	// Signal Channel to Update Scheduler Class about changes
	updateSignal chan updateSignalOp

	// SignalChan for termination
	stopScheduleSignal chan interface{}

	// Concurrent safe JobMap
	activeJobs jobMap

	// Wait-group
	wg sync.WaitGroup

	// Logging Interface
	Logger logr.Logger

	// Lock the Schedule Structure for Modifications
	mx deadlock.RWMutex

	// Middleware to run
	executationMiddleWares []ExecutionMiddleWare

	// Retry Middleware to run
	retryMiddlewares []RetryMiddleware

	// Context for Jobs
	Ctx context.Context
}

// NewSchedule Create a new schedule for` jobFunc func()` that will run according to `timer Timer` with the supplied []Options
func NewSchedule(ctx context.Context, id string, timer Timer, jobFunc func(context.Context), opts ...Option) *Task {
	var options = defaultTaskOptions()

	// Apply Options
	for _, option := range opts {
		option.apply(options)
	}

	s := &Task{
		id:                     id,
		jobSrcFunc:             jobFunc,
		timer:                  timer,
		activeJobs:             *newJobMap(),
		Logger:                 options.logger.WithValues("taskid", id),
		executationMiddleWares: options.executationmiddlewares,
		retryMiddlewares:       options.retryMiddlewares,
		Ctx:                    ctx,
	}
	t, _ := timer.Next()
	s.nextRun.Set(t)
	return s
}

func (s *Task) GetID() string {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.id
}

// Start Start the scheduler. Method is concurrent safe. Calling Start() have the following effects according to the
//	scheduler state:
//		1. NEW: Start the Schedule; running the defined Job on the first Timer's Next() time.
//		2. QUEUED: No Effect (and prints warning)
//		3. STOPPED: Restart the schedule
//		4. FINISHED: No Effect (and prints warning)
func (s *Task) Start() {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.Logger.Info("Job Schedule Started")
	metrics.SetGaugeWithLabels(schedmetrics.GetMetricsGaugeKey(schedmetrics.Metrics_Guage_Up), 1, []metrics.Label{{Name: "id", Value: s.id}})

	// Create stopSchedule signal channel, buffer = 1 to allow non-blocking signaling.
	s.stopScheduleSignal = make(chan interface{}, 1)

	for _, mw := range s.executationMiddleWares {
		s.Logger.V(1).Info("Initilized Executation Middleware", "middleware", mw)
		mw.Initilize(s)
	}
	for _, mw := range s.retryMiddlewares {
		s.Logger.V(1).Info("Initilized Retry Middleware", "middleware", mw)
		mw.Initilize(s)
	}

	//go s.scheduleLoop()
	//go func() {}()
}

// Stop stops the scheduler. Method is **Blocking** and concurrent safe. When called:
//		1. Schedule will cancel all waiting scheduled jobs.
//		2. Schedule will wait for all running jobs to finish.
//	Calling Stop() has the following effects depending on the state of the schedule:
//		1. NEW: No Effect
//		2. QUEUED: Stop Schedule
//		3. STOPPED: No Effect
//		4. FINISHED: No Effect
func (s *Task) Stop() {
	s.mx.Lock()
	defer s.mx.Unlock()

	// Stop control loop
	s.Logger.Info("Stopping Schedule...")
	s.stopScheduleSignal <- struct{}{}

	// Print No. of Active Jobs
	if noOfActiveJobs := s.activeJobs.len(); s.activeJobs.len() > 0 {
		s.Logger.Info("Waiting for active jobs still running...", "jobs", noOfActiveJobs)
	}

	s.wg.Wait()
	s.Logger.Info("Job Schedule Stopped")
	metrics.SetGaugeWithLabels(schedmetrics.GetMetricsGaugeKey(schedmetrics.Metrics_Guage_Up), 0, []metrics.Label{{Name: "id", Value: s.id}})
	close(s.stopScheduleSignal)
}

func (s *Task) runPreExecutationMiddlware() (MWResult, error) {
	for _, middleware := range s.executationMiddleWares {
		s.Logger.V(1).Info("Running Handler", "middleware", middleware)
		metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_PreExecutationRuns), 1, []metrics.Label{{Name: "id", Value: s.id}, {Name: "middleware", Value: fmt.Sprintf("%T", middleware)}})
		result, err := middleware.PreHandler(s)
		if err != nil {
			s.Logger.Error(err, "Middleware Returned Error", "middleware", middleware, "result", result)
		} else {
			s.Logger.Info("Middleware Returned No Error", "middleware", middleware, "result", result)
		}

		switch result.Result {
		case MWResult_Defer:
			metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_DeferredJobs), 1, []metrics.Label{{Name: "id", Value: s.id}})
			return result, err
		case MWResult_NextMW:
			continue
		case MWResult_Cancel:
			return result, err
		}
	}
	return MWResult{Result: MWResult_NextMW}, nil
}

func (s *Task) runRetryMiddleware(prerun bool, err error) {
	for _, retrymiddleware := range s.retryMiddlewares {
		s.Logger.V(1).Info("Running Retry Middleware", "middleware", retrymiddleware)
		metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_PreRetryRuns), 1, []metrics.Label{{Name: "id", Value: s.id}, {Name: "middleware", Value: fmt.Sprintf("%T", retrymiddleware)}, {Name: "Prerun", Value: strconv.FormatBool(prerun)}})

		retryops, _ := retrymiddleware.Handler(s, prerun, err)

		switch retryops.Result {
		case RetryResult_Retry:
			s.Logger.V(1).Info("Retry Middleware Delayed Job", "middleware", retrymiddleware, "duration", retryops.Delay)
			metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_PreRetryRetries), 1, []metrics.Label{{Name: "id", Value: s.id}, {Name: "middleware", Value: fmt.Sprintf("%T", retrymiddleware)}, {Name: "Prerun", Value: strconv.FormatBool(prerun)}})
			s.retryJob(retryops.Delay)
		case RetryResult_NoRetry:
			s.Logger.V(1).Info("Retry Middleware Canceled Retries", "middleware", retrymiddleware)
			metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_PreRetryResets), 1, []metrics.Label{{Name: "id", Value: s.id}, {Name: "middleware", Value: fmt.Sprintf("%T", retrymiddleware)}, {Name: "Prerun", Value: strconv.FormatBool(prerun)}})
		case RetryResult_NextMW:
			s.Logger.V(1).Info("Retry Middleware Skipped", "middleware", retrymiddleware)
			metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_PreRetrySkips), 1, []metrics.Label{{Name: "id", Value: s.id}, {Name: "middleware", Value: fmt.Sprintf("%T", retrymiddleware)}, {Name: "Prerun", Value: strconv.FormatBool(prerun)}})
		}
	}
}

func (s *Task) runPostExecutionHandler(err error) MWResult {
	for _, postmiddleware := range s.executationMiddleWares {
		s.Logger.V(1).Info("Running PostHandler Middlware", "middleware", postmiddleware)
		metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_PostExecutationFailedRuns), 1, []metrics.Label{{Name: "id", Value: s.id}, {Name: "middleware", Value: fmt.Sprintf("%T", postmiddleware)}})
		result := postmiddleware.PostHandler(s, err)
		switch result.Result {
		case MWResult_Defer:
			return MWResult{Result: MWResult_Defer}
		case MWResult_Cancel:
			return MWResult{Result: MWResult_Cancel}
		case MWResult_NextMW:
			continue
		}
	}

	return MWResult{Result: MWResult_NextMW}
}

func (s *Task) runJobInstance(result chan interface{}) {
	s.wg.Add(1)
	defer s.wg.Done()

	// Create a new instance of s.jobSrcFunc
	jobInstance := job.NewJob(s.Ctx, s.jobSrcFunc)

	joblog := s.Logger.WithValues("instance", jobInstance.ID())
	joblog.V(1).Info("Job Run Starting")

	// Add to active jobs map
	s.activeJobs.add(jobInstance)
	defer s.activeJobs.delete(jobInstance)

	// Logs and Metrics --------------------------------------
	// -------------------------------------------------------
	labels := []metrics.Label{{Name: "id", Value: s.id}}
	metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_Runs), 1, labels)
	if s.activeJobs.len() > 1 {
		metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_OverlappingRuns), 1, labels)
	}
	// -------------------------------------------------------

	// Synchronously Run Job Instance
	lastError := jobInstance.Run()

	// -------------------------------------------------------
	// Logs and Metrics --------------------------------------
	if lastError != nil {
		joblog.
			WithValues("duration", jobInstance.ActualElapsed().Round(1*time.Millisecond)).
			WithValues("state", jobInstance.State().String()).
			WithValues("error", lastError.Error()).
			Error(lastError, "Job Error")
		metrics.IncrCounterWithLabels([]string{"sched", "runerrors"}, 1, labels)
		result <- lastError
	} else {
		joblog.
			WithValues("duration", jobInstance.ActualElapsed().Round(1*time.Millisecond)).
			WithValues("state", jobInstance.State().String()).
			Info("Job Finished")
		result <- nil
	}
}

func negativeToZero(nextRunDuration time.Duration) time.Duration {
	if nextRunDuration < 0 {
		nextRunDuration = 0
	}
	return nextRunDuration
}

func (s *Task) retryJob(in time.Duration) {
	s.Logger.
		WithValues("duration", in).
		V(1).Info("Rescheduling Job")
	s.timer.Reschedule(in)
}

func (s *Task) GetNextRun() time.Time {
	return s.nextRun.Get()
}

func (s *Task) Run() {
	jobResultSignal := make(chan interface{})
	defer close(jobResultSignal)
	s.Logger.Info("Checking Pre Execution Middleware")
	result, err := s.runPreExecutationMiddlware()
	switch result.Result {
	case MWResult_Cancel:
		s.Logger.Info("Scheduled Job run is Canceled")
		t, _ := s.timer.Next()
		s.nextRun.Set(t)
		s.sendUpdateSignal(updateSignalOp_Reschedule)
		return
	case MWResult_Defer:
		s.Logger.Info("Scheduled Job will be Retried")
		s.runRetryMiddleware(true, err)
		t, _ := s.timer.Next()
		s.nextRun.Set(t)
		s.sendUpdateSignal(updateSignalOp_Reschedule)
		return
	case MWResult_NextMW:
		s.Logger.Info("Dispatching Job")
		go s.runJobInstance(jobResultSignal)
		t, _ := s.timer.Next()
		s.nextRun.Set(t)
		s.sendUpdateSignal(updateSignalOp_Reschedule)
	}
	select {
	case result := <-jobResultSignal:
		s.Logger.
			WithValues("result", result).
			Info("Got Job Result", "result", result)
		err, ok := result.(joberrors.FailedJobError)
		if ok {
			s.Logger.Error(err, "Job Failed")

			metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_FailedJobs), 1, []metrics.Label{{Name: "id", Value: s.id}})

			mwresult := s.runPostExecutionHandler(err)

			if mwresult.Result == MWResult_Defer {
				/* run Retry Framework */
				s.Logger.V(1).Info("Post Executation Middleware Retry Request")
				s.runRetryMiddleware(false, err)
			}
		} else {
			metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_SucceededJobs), 1, []metrics.Label{{Name: "id", Value: s.id}})
			s.runPostExecutionHandler(nil)
		}
	}
	t, _ := s.timer.Next()
	s.nextRun.Set(t)
	s.sendUpdateSignal(updateSignalOp_Reschedule)
}

func (s *Task) sendUpdateSignal(op UpdateSignalOp_Type) {
	s.Logger.V(1).Info("Sending Update Signal")
	s.updateSignal <- updateSignalOp{id: s.id, operation: op}
	s.Logger.V(1).Info("Sent Update Signal")
}
