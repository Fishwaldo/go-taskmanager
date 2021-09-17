package taskmanager

import (
	"fmt"
	"time"

	"github.com/gorhill/cronexpr"
)

//Timer is an Interface for a Timer object that is used by a Schedule to determine when to run the next run of a job.
// Timer need to implement the Next() method returning the time of the next Job run. Timer indicates that no jobs shall
// be scheduled anymore by returning done == true. The `next time.Time` returned with `done bool` == true IS IGNORED.
// Next() shall not return time in the past. Time in the past is reset to time.Now() at evaluation time in the scheduler.
type Timer interface {
	Next() (next time.Time, done bool)
	Reschedule(delay time.Duration)
}

//Once A timer that run ONCE after an optional specific delay.
type Once struct {
	delay time.Duration
	done  bool
}

//NewOnce Return a timer that trigger ONCE after `d` delay as soon as Timer is inquired for the next Run.
//Delay = 0 means the Timer return now(), aka as soon as time is inquired.
func NewOnce(d time.Duration) (*Once, error) {
	if d < 0 {
		return nil, fmt.Errorf("invalid d, must be >= 0")
	}
	return &Once{
		delay: d,
	}, nil
}

// NewOnceTime Return a timer that trigger ONCE at `t` time.Time.
//If `t` is in the past at inquery time, timer will NOT run.
func NewOnceTime(t time.Time) (*Once, error) {
	remaining := time.Until(t)
	if remaining < 0 {
		return &Once{
			delay: remaining,
			done:  true,
		}, nil
	}
	return &Once{
		delay: remaining,
	}, nil
}

//Next Return Next Time OR a boolean indicating no more Next()(s)
func (o *Once) Next() (time.Time, bool) {
	if !o.done {
		o.done = true
		return time.Now().Add(o.delay), false
	}
	return time.Time{}, o.done
}

func (o *Once) Reschedule(d time.Duration) {
	o.delay = d
	if o.done {
		o.done = false
	}
}

//Fixed A Timer that fires at a fixed duration intervals
type Fixed struct {
	duration time.Duration
	next     time.Time
	delay    time.Duration
}

//NewFixed Returns Fixed Timer; A Timer that fires at a fixed duration intervals.
func NewFixed(duration time.Duration) (*Fixed, error) {
	if duration < 0 {
		return nil, fmt.Errorf("invalid duration, must be >= 0")
	}
	return &Fixed{
		duration: duration,
		next:     time.Now().Add(duration),
	}, nil
}

//Next Return Next fire time.
func (f *Fixed) Next() (time.Time, bool) {
	now := time.Now()
	if f.delay > 0 {
		next := now.Add(f.delay)
		f.delay = 0
		return next, false
	}
	f.next = time.Now().Add(f.duration)
	return f.next, false
}

func (f *Fixed) Reschedule(t time.Duration) {
	f.delay = t
}

//Cron A Timer that fires at according to a cron expression.
//All expresion supported by `https://github.com/gorhill/cronexpr` are supported.
type Cron struct {
	expression cronexpr.Expression
	delay      time.Duration
}

//NewCron returns a Timer that fires at according to a cron expression.
//All expresion supported by `https://github.com/gorhill/cronexpr` are supported.
func NewCron(cronExpression string) (*Cron, error) {
	expression, err := cronexpr.Parse(cronExpression)
	if err != nil {
		return nil, fmt.Errorf("cron expression invalid: %w", err)
	}
	return &Cron{expression: *expression}, nil
}

//Next Return Next fire time.
func (c *Cron) Next() (time.Time, bool) {
	if c.delay > 0 {
		next := time.Now().Add(c.delay)
		c.delay = 0
		return next, false
	}
	return c.expression.Next(time.Now()), false
}

func (c *Cron) Reschedule(d time.Duration) {
	c.delay = d
}
