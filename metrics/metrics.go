package schedmetrics

import (
//	"github.com/armon/go-metrics/prometheus"
)

const (
	Metrics_Guage_Up = iota
	Metrics_Guage_Jobs
)

const (
	Metrics_Counter_ContextCancels = iota
	Metrics_Counter_Stops
	Metrics_Counter_Reschedule
	Metrics_Counter_FailedJobs
	Metrics_Counter_PostExecutationFailedRuns
	Metrics_Counter_PostExecutationRetriesRequested
	Metrics_Counter_PostExecutationRetryRuns
	Metrics_Counter_PostExecutationRetries
	Metrics_Counter_PostExecutationRetryResets
	Metrics_Counter_PostExecutationSkips
	Metrics_Counter_PostExecucutionDefaultRetries
	Metrics_Counter_SucceededJobs
	Metrics_Counter_PostExecutationRuns
	Metrics_Counter_PreExecutationRuns
	Metrics_Counter_DeferredJobs
	Metrics_Counter_PreRetryRuns
	Metrics_Counter_PreRetryRetries
	Metrics_Counter_PreRetryResets
	Metrics_Counter_PreRetrySkips
	Metrics_Counter_PreRetryDefault
	Metrics_Counter_Runs
	Metrics_Counter_OverlappingRuns
	Metrics_Counter_Runerrors
	Metrics_Counter_MW_ConcurrentJob_Blocked
	Metrics_Counter_MW_HasTags_Blocked
	Metrics_Counter_MW_ConstantBackoff_Retries
	Metrics_Counter_MW_ExpBackoff_Retries
	Metrics_Counter_MW_RetryLimit_Hit
)

type GaugeValues struct {
	Name []string
	Help string
}

type CounterValues struct {
	Name []string
	Help string
}

type SummaryValues struct {
	Name []string
	Help string
}


var MetricsGauges = func() map[int]GaugeValues {
	return map[int]GaugeValues {
	Metrics_Guage_Up: 
		{
			Name: []string{"sched", "up"},
			Help: "If the Job Scheduler is Active",
		},
	Metrics_Guage_Jobs:
		{
			Name: []string{"sched", "jobs"},
			Help: "Number of Jobs Scheduled",
		},
	}
}

var MetricsCounter = func() map[int]CounterValues {
	return map[int]CounterValues {
	Metrics_Counter_ContextCancels:
		{
			Name: []string{"sched", "contextcancels"},
			Help: "Number of Times a Job was canceled by a Context",
		},
	Metrics_Counter_Stops:
		{
			Name: []string{"sched", "stops"},
			Help: "Number of Times a Job was Stopped",
		},
	Metrics_Counter_Reschedule:
		{
			Name: []string{"sched", "reschedules"},
			Help: "Number of times a Job was rescheduled",
		},
	Metrics_Counter_FailedJobs:
		{
			Name: []string{"sched", "failedjobs"},
			Help: "Number of Failed Job Runs",
		},
	Metrics_Counter_PostExecutationFailedRuns:
		{
			Name: []string{"sched", "postexecutationfailedmiddlewareruns"},
			Help: "Number of Post Exececutation Failure Middlwares were executed",
		},
	Metrics_Counter_PostExecutationRetriesRequested:
		{
			Name: []string{"sched", "postexecutionfailureretriesrequested"},
			Help: "Number of Post Exececution Retries Requested",
		},
	Metrics_Counter_PostExecutationRetryRuns:
		{
			Name: []string{"sched", "postretrymiddlewareruns"},
			Help: "Number of Post Failure Retry Middlewwares executed",
		},
	Metrics_Counter_PostExecutationRetries:
		{
			Name: []string{"sched", "postretrymiddlewareretries"},
			Help: "Number of Post Failure Retries executed with a Retry time set",
		},
	Metrics_Counter_PostExecutationRetryResets:
		{
		Name: []string{"sched", "postretrymiddlewareresets"},
		Help: "Number of Post Failure Retry Middlewares that were Reset",
		},
	Metrics_Counter_PostExecutationSkips:
		{
			Name: []string{"sched", "postretrymiddlewareskips"},
			Help: "Number of Post Failure Retry Middlware that skipped setting a duration",
		},
	Metrics_Counter_PostExecucutionDefaultRetries:
		{
			Name: []string{"sched", "postretrymiddlewaredefaultretry"},
			Help: "Number of Default Retry Durations executed During Post Executation stage - Failures",
		},
	Metrics_Counter_SucceededJobs:
		{
		Name: []string{"sched", "succceededjobs"},
		Help: "Number of Successful Job Runs",
		},		
	Metrics_Counter_PostExecutationRuns:
		{
			Name: []string{"sched", "postexecutationsucceededmiddlewareruns"},
			Help: "Number of Times a Post Executation Middlware was run after a successful return",
		},
	Metrics_Counter_PreExecutationRuns:
		{
			Name: []string{"sched", "preexecutationmiddlewareruns"},
			Help: "Number of times Pre Executation Middlewares were run",
		},
	Metrics_Counter_DeferredJobs:
		{
			Name: []string{"sched", "deferredjobs"},
			Help: "Number of times jobs were deferred",
		},
	Metrics_Counter_PreRetryRuns:
		{
		Name: []string{"sched", "preretrymiddlewareruns"},
		Help: "Number of times Pre-Retry Middlewares were run",
		},
	Metrics_Counter_PreRetryRetries:
		{
			Name: []string{"sched", "preretrymiddlewareretries"},
			Help: "Number of Pre Retry Middlewares that Set a Retry Duration",
		},
	Metrics_Counter_PreRetryResets:
		{
		Name: []string{"sched", "preretrymiddlewareresets"},
		Help: "Number of Pre Retry Middlewares that were Reset",
		},
	Metrics_Counter_PreRetrySkips:
		{
			Name: []string{"sched", "preretrymiddlewareskips"},
			Help: "Number of Pre Retry Middlewares that Skipped Setting a Retry Duration",
		},		
	Metrics_Counter_PreRetryDefault:
		{
			Name: []string{"sched", "preretrymiddlewaredefaultretry"},
			Help: "Number of Default Retry Durations executed During Pre-Executation stage",
		},	
	Metrics_Counter_Runs:
		{
			Name: []string{"sched", "runs"},
			Help: "Number of Times a Job has run",
		},
	Metrics_Counter_OverlappingRuns:
		{
			Name: []string{"sched", "overlappingRuns"},
			Help: "Number of Times a Job has Overlapped with a previous Job",
		},
	Metrics_Counter_Runerrors:
		{
			Name: []string{"sched", "runerrors"},
			Help: "Number of Times a Job Errored Out",
		},
	Metrics_Counter_MW_ConcurrentJob_Blocked:
		{
			Name: []string{"sched", "middleware", "concurrentjob", "blocked"},
			Help: "Number of Times a Job was Blocked from running",
		},
	Metrics_Counter_MW_HasTags_Blocked:
		{
			Name: []string{"sched", "middleware", "hastags", "blocked"},
			Help: "Number of times the Has Tags Middlware Blocked Executation",
		},
	Metrics_Counter_MW_ConstantBackoff_Retries:
		{
			Name: []string{"sched", "middleware", "constantbackoff", "retries"},
			Help: "Number of times the ConstantBackoff Middleware set a Retry Time",
		},
	Metrics_Counter_MW_ExpBackoff_Retries:
		{
			Name: []string{"sched", "middleware", "exponentialbackoff", "retries"},
			Help: "Number of times the Exponential Backoff Middleware set a Retry Time",
		},
	Metrics_Counter_MW_RetryLimit_Hit:
		{
			Name: []string{"sched", "middleware", "retrylimit", "hit"},
			Help: "Number of times the Retry Limit Middleware Canceled a pending job",
		},

	}
}

var MetricsSummary = func() map[int]SummaryValues {
	return map[int]SummaryValues {
	}
}



func GetMetricsGaugeKey(key int) []string {
	return MetricsGauges()[key].Name
}

func GetMetricsCounterKey(key int) []string {
	return MetricsCounter()[key].Name
}

func GetMetricsSummaryKey(key int) []string {
	return MetricsSummary()[key].Name
}