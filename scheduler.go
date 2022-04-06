package taskmanager

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/sasha-s/go-deadlock"
	"github.com/go-logr/logr"
	"github.com/Fishwaldo/go-taskmanager/joberrors"
	schedmetrics "github.com/Fishwaldo/go-taskmanager/metrics"
	"github.com/armon/go-metrics"
)

// Scheduler manage one or more Schedule creating them using common options, enforcing unique IDs, and supply methods to
// Start / Stop all schedule(s).
type Scheduler struct {
	tasks              map[string]*Task
	nextRun            timeSlice
	mx                 deadlock.RWMutex
	tsmx               deadlock.RWMutex
	log                logr.Logger
	updateScheduleChan chan updateSignalOp
	scheduleOpts       []Option
}

type UpdateSignalOp_Type int

const (
	updateSignalOp_Reschedule UpdateSignalOp_Type = iota
)

type updateSignalOp struct {
	id        string
	operation UpdateSignalOp_Type
}

type timeSlice []*Task

func (p timeSlice) Len() int {
	return len(p)
}

func (p timeSlice) Less(i, j int) bool {
	if p[i].GetNextRun().IsZero() {
		return false
	}
	return p[i].GetNextRun().Before(p[j].GetNextRun())
}

func (p timeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

//NewScheduler Creates new Scheduler, opt Options are applied to *every* schedule added and created by this scheduler.
func NewScheduler(opts ...Option) *Scheduler {
	var options = defaultSchedOptions()

	// Apply Options
	for _, option := range opts {
		option.apply(options)
	}

	s := &Scheduler{
		tasks:              make(map[string]*Task),
		nextRun:            make(timeSlice, 0),
		updateScheduleChan: make(chan updateSignalOp, 100),
		scheduleOpts:       opts,
		log:                options.logger,
	}

	go s.scheduleLoop()
	return s
}

//Add Create a new Task for` jobFunc func()` that will run according to `timer Timer` with the []Options of the Scheduler.
func (s *Scheduler) Add(ctx context.Context, id string, timer Timer, job func(context.Context), extraOpts ...Option) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if _, ok := s.tasks[id]; ok {
		return joberrors.ErrorScheduleExists{Message: "job with this id already exists"}
	}

	// Create schedule
	opts := append(extraOpts, s.scheduleOpts...)
	schedule := NewSchedule(ctx, id, timer, job, opts...)
	schedule.updateSignal = s.updateScheduleChan
	// Add to managed schedules
	s.tasks[id] = schedule
	metrics.SetGauge(schedmetrics.GetMetricsGaugeKey(schedmetrics.Metrics_Guage_Jobs), float32(len(s.tasks)))

	s.log.Info("Added New Job", "jobid", schedule.GetID())
	return nil
}

//Start Start the Schedule with the given id. Return error if no Schedule with the given id exist.
func (s *Scheduler) Start(id string) error {
	s.mx.Lock()

	// Find Schedule by id
	schedule, found := s.tasks[id]
	if !found {
		return joberrors.ErrorScheduleNotFound{Message: "Schedule Not Found"}
	}

	// Start it ¯\_(ツ)_/¯
	schedule.Start()
	s.mx.Unlock()
	s.addScheduletoRunQueue(schedule)
	s.log.Info("Start Job", "jobid", schedule.GetID())
	return nil
}

//StartAll Start All Schedules managed by the Scheduler
func (s *Scheduler) StartAll() {
	s.log.Info("StartAll Called")
	for _, schedule := range s.tasks {
		s.Start(schedule.id)
	}
}

//Stop Stop the Schedule with the given id. Return error if no Schedule with the given id exist.
func (s *Scheduler) Stop(id string) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	schedule, found := s.tasks[id]
	if !found {
		return joberrors.ErrorScheduleNotFound{Message: "Schedule Not Found"}
	}
	s.tsmx.Lock()
	for pos, sched := range s.nextRun {
		if sched.id == id {
			s.nextRun = append(s.nextRun[:pos], s.nextRun[pos+1])
			break
		}
	}
	s.tsmx.Unlock()
	schedule.Stop()
	return nil
}

//StopAll Stops All Schedules managed by the Scheduler concurrently, but will block until ALL of them have stopped.
func (s *Scheduler) StopAll() {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.tsmx.Lock()
	s.nextRun = nil
	s.tsmx.Unlock()
	wg := sync.WaitGroup{}
	wg.Add(len(s.tasks))
	for _, schedule := range s.tasks {
		go func(scheduleCpy *Task) {
			scheduleCpy.Stop()
			wg.Done()
		}(schedule)
	}
	wg.Wait()
}

//GetSchedule Returns a Schedule by ID from the Scheduler
func (s *Scheduler) GetSchedule(id string) (*Task, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	j, ok := s.tasks[id]
	if !ok {
		return nil, joberrors.ErrorScheduleNotFound{Message: "Schedule Not Found"}
	}
	return j, nil
}

//GetAllSchedules Returns all Schedule's in the Scheduler
func (s *Scheduler) GetAllSchedules() (map[string]*Task, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.tasks, nil
}

func (s *Scheduler) getNextJob() *Task {
	s.tsmx.RLock()
	defer s.tsmx.RUnlock()
	if s.nextRun.Len() == 0 {
		return nil
	}
	for _, sched := range s.nextRun {
		if sched.nextRun.Get().IsZero() {
			s.log.Info("NextRun is Zero", "jobid", sched.GetID())
			continue
		}
		return sched
	}
	return nil
}

func (s *Scheduler) scheduleLoop() {
	var nextRun time.Time
	var nextRunChan <-chan time.Time
	for {
		nextjob := s.getNextJob()
		if nextjob != nil {
			nextRun = nextjob.GetNextRun()
			s.log.Info("Next Scheduler Run", "next", time.Until(nextRun), "jobid", nextjob.GetID())
			nextRunChan = time.After(time.Until(nextRun))
		} else {
			s.log.Info("No Jobs Scheduled")
		}

		select {
		case <-nextRunChan:
			if nextjob != nil {
				s.log.Info("Dispatching Job", "jobid", nextjob.id)
				nextjob.nextRun.Set(time.Time{})
				go nextjob.Run()
			} else {
				s.log.Error(nil, "nextjob is Nil")
			}
		case op := <-s.updateScheduleChan:
			switch op.operation {
			case updateSignalOp_Reschedule:
				s.log.Info("recalcSchedule Triggered", "operation", op.id)
				s.updateNextRun()
			default:
				s.log.Error(nil, "Unhandled updateSignalOp Recieved")
			}
		}
	}

}

func (s *Scheduler) updateNextRun() {
	s.tsmx.Lock()
	defer s.tsmx.Unlock()
	sort.Sort(s.nextRun)
	for _, job := range s.nextRun {
		s.log.Info("Next Run", "jobid", job.GetID(), "when", job.GetNextRun().Format(time.RFC1123))
	}
}

func (s *Scheduler) addScheduletoRunQueue(schedule *Task) {
	s.tsmx.Lock()
	defer s.tsmx.Unlock()
	s.nextRun = append(s.nextRun, schedule)
	s.log.Info("addScheduletoRunQueue", "jobid", schedule.GetID())
	for _, job := range s.nextRun {
		s.log.Info("Job Run Queue", "jobid", job.GetID(), "when", job.GetNextRun().Format(time.RFC1123))
	}
	s.updateScheduleChan <- updateSignalOp{operation: updateSignalOp_Reschedule, id: schedule.id}
}
