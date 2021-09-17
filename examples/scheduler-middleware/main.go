package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fishwaldo/go-taskmanager"
	"github.com/Fishwaldo/go-taskmanager/job"
	logruslog "github.com/Fishwaldo/go-taskmanager/loggers/logrus"
	prometheusConfig "github.com/Fishwaldo/go-taskmanager/metrics/prometheus"
	executionmiddleware "github.com/Fishwaldo/go-taskmanager/middleware/executation"
	retrymiddleware "github.com/Fishwaldo/go-taskmanager/middleware/retry"

	//"github.com/armon/go-metrics"
	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := prometheus.PrometheusOpts{
		GaugeDefinitions:   prometheusConfig.GetPrometicsGaugeConfig(),
		CounterDefinitions: prometheusConfig.GetPrometicsCounterConfig(),
		SummaryDefinitions: prometheusConfig.GetPrometicsSummaryConfig(),
	}

	sink, err := prometheus.NewPrometheusSinkFrom(cfg)
	if err != nil {
		log.Fatal("Cant Create Prometheus Sink")
	}
	defer prom.Unregister(sink)

	metrics.NewGlobal(metrics.DefaultConfig("test"), sink)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	exbomw := retrymiddleware.NewDefaultRetryExponentialBackoff()

	thmw := executionmiddleware.NewTagHandler()
	thmw.SetRequiredTags("Hello")

	cjl := executionmiddleware.NewCJLock()
	mrt := retrymiddleware.NewRetryRetryCountLimit(5)

	job1 := func(seconds time.Duration) func(context.Context) {
		return func(ctx context.Context) {
			jobrunner, _ := ctx.Value(job.JobCtxValue{}).(*job.Job)
			log.Println("Job ", jobrunner.ID(), " Duration: ", seconds*time.Second, "\t Doing some work...")
			if thmw.IsHaveTag("Hello") {
				thmw.DelHaveTags("Hello")
			} else {
				thmw.SetHaveTags("Hello")
			}
			select {
			case <-ctx.Done():
				log.Println("Job ", jobrunner.ID(), " Context Cancelled Job")
			case <-time.After(time.Second * seconds):
				log.Println("Job ", jobrunner.ID(), " Work Done")
			}
			//log.Panic("Job ", jobrunner.ID(), "Pannic Test")
			log.Println("Job ", jobrunner.ID(), "Duration: ", seconds*time.Second, "\t Finished Work.")
		}
	}

	job2 := func(seconds time.Duration) func(context.Context) {
		return func(ctx context.Context) {
			jobrunner, _ := ctx.Value(job.JobCtxValue{}).(*job.Job)
			select {
			case <-ctx.Done():
				log.Println("NeedTagsJob ", jobrunner.ID(), " Context Cancelled Job")
			default:
				log.Println("NeedTagsJob ", jobrunner.ID(), "Is Running")
			}
		}
	}

	cronTimer, err := taskmanager.NewCron("* * * * *")
	if err != nil {
		panic(fmt.Sprintf("invalid cron expression: %s", err.Error()))
	}

	cronTimer5, err := taskmanager.NewCron("*/1 * * * *")
	if err != nil {
		panic(fmt.Sprintf("invalid cron expression: %s", err.Error()))
	}

	fixedTimer10second, err := taskmanager.NewFixed(10 * time.Second)
	if err != nil {
		panic(fmt.Sprintf("invalid interval: %s", err.Error()))
	}

	onceAfter10s, err := taskmanager.NewOnce(10 * time.Second)
	if err != nil {
		panic(fmt.Sprintf("invalid delay: %s", err.Error()))
	}

	//	logger := sched.DefaultLogger()
	//	logger.SetLevel(sched.LOG_TRACE)

	logger := logruslog.LogrusDefaultLogger()

	// Create Schedule
	scheduler := taskmanager.NewScheduler(
		taskmanager.WithLogger(logger),
	)

	//ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())

	_ = cronTimer
	_ = onceAfter10s
	_ = cronTimer5
	//_ = scheduler.Add(ctx1, "cronEveryMinute", cronTimer, job1(12))
	//_ = scheduler.Add(ctx2, "cronEvery5Minute", cronTimer5, job1(8))
	//_ = scheduler.Add(ctx1, "fixedTimer10second", cronTimer, job1(1))
	_ = scheduler.Add(ctx2, "fixedTimer10second30SecondDuration", fixedTimer10second, job1(21), taskmanager.WithExecutationMiddleWare(cjl), taskmanager.WithRetryMiddleWare(mrt), taskmanager.WithRetryMiddleWare(exbomw))
	_ = scheduler.Add(ctx2, "TagHandler", fixedTimer10second, job2(5), taskmanager.WithExecutationMiddleWare(thmw))
	//_ = scheduler.Add(ctx2, "onceAfter10s", onceAfter10s, job1(12))

	scheduler.StartAll()
	//scheduler.Start("TagHandler")

	// Listen to CTRL + C
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	_ = <-signalChan

	// Send Cancel Signals to our Jobs
	//cancel1()
	cancel2()

	// Stop before shutting down.
	scheduler.StopAll()

}
