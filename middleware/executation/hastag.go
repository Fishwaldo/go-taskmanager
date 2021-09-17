package executionmiddleware

import (
	"sync"

	"github.com/Fishwaldo/go-taskmanager"
	"github.com/Fishwaldo/go-taskmanager/joberrors"
	schedmetrics "github.com/Fishwaldo/go-taskmanager/metrics"
	"github.com/armon/go-metrics"
)

var _ taskmanager.ExecutionMiddleWare = (*HasTagHandler)(nil)

type hasTagCtxKey struct{}

// HasTagHandler is a Middleware that will Defer jobs if the Requirements are not meet
// Requirements are Specified as "Tags" (strings) and a Job has a list of Tags needed
// When a Job is about to be dispatched, the Jobs "Required" tags are compared against a
// list of "available" tags, and if the available tags does not match the required tags, the
// job is deferred (or canceled if there is no other Middleware after this one.)
type HasTagHandler struct {
	mx           sync.RWMutex
	requiredTags map[string]bool
	haveTags     map[string]bool
}

// SetHaveTags Set a tag indicating that a resource is available. Before a job runs with a Matching
// Required tag, it checks if its present.
func (hth *HasTagHandler) SetHaveTags(tag string) {
	hth.mx.Lock()
	defer hth.mx.Unlock()
	hth.haveTags[tag] = true
}

// DelHaveTags Delete a tag indicating a resource is no longer available.
func (hth *HasTagHandler) DelHaveTags(tag string) {
	hth.mx.Lock()
	defer hth.mx.Unlock()
	delete(hth.haveTags, tag)
}

// IsHaveTag Test if a resource represented by tag is present
func (hth *HasTagHandler) IsHaveTag(tag string) bool {
	hth.mx.RLock()
	defer hth.mx.RUnlock()
	_, ok := hth.haveTags[tag]
	return ok
}

// SetRequiredTags add a tag that makes a job dependant upon a resource being available (set via SetHaveTag)
func (hth *HasTagHandler) SetRequiredTags(tag string) {
	hth.mx.Lock()
	defer hth.mx.Unlock()
	hth.requiredTags[tag] = true
}

// DelRequiredTags delete a tag that a job requires.
func (hth *HasTagHandler) DelRequiredTags(tag string) {
	hth.mx.Lock()
	defer hth.mx.Unlock()
	delete(hth.requiredTags, tag)
}

// IsRequiredTag test if a resource requirement represented by tag is available
func (hth *HasTagHandler) IsRequiredTag(tag string) bool {
	hth.mx.RLock()
	defer hth.mx.RUnlock()
	_, ok := hth.requiredTags[tag]
	return ok
}

// Handler Runs the Tag Handler before a job is dispatched.
func (hth *HasTagHandler) PreHandler(s *taskmanager.Task) (taskmanager.MWResult, error) {
	s.Logger.
		With("Present", hth.haveTags).
		With("Required", hth.requiredTags).
		Debug("Checking Tags")
	for k := range hth.requiredTags {
		if !(hth.IsHaveTag(k)) {
			s.Logger.
				With("Tag", k).
				Warn("Missing Tag")
			metrics.IncrCounterWithLabels(schedmetrics.GetMetricsCounterKey(schedmetrics.Metrics_Counter_MW_HasTags_Blocked), 1, []metrics.Label{{Name: "id", Value: s.GetID()}})
			return taskmanager.MWResult{Result: taskmanager.MWResult_Defer}, joberrors.FailedJobError{Message: "Missing Tag", ErrorType: joberrors.Error_DeferedJob}
		} else {
			s.Logger.
				With("Tag", k).
				Info("Tag Present")
		}
	}
	return taskmanager.MWResult{Result: taskmanager.MWResult_NextMW}, nil
}

func (hth *HasTagHandler) PostHandler(s *taskmanager.Task, err error) taskmanager.MWResult {
	return taskmanager.MWResult{Result: taskmanager.MWResult_NextMW}
}

func (hth *HasTagHandler) Initilize(s *taskmanager.Task) {

}

func (hth *HasTagHandler) Reset(s *taskmanager.Task) {

}

// NewTagHandlerMW Create a new Tag Handler Middleware
func NewTagHandler() *HasTagHandler {
	val := HasTagHandler{
		requiredTags: make(map[string]bool),
		haveTags:     make(map[string]bool),
	}
	return &val
}
