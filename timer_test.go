package taskmanager

import (
	"testing"
	"time"
)

func TestTimerOnce(t *testing.T) {
	timer, err := NewOnce(1 * time.Second)
	if err != nil {
		t.Errorf("NewOnce Timer Returned Error %s", err.Error())
	}
	next, run := timer.Next()
	if !next.Round(time.Second).Equal(time.Now().Add(1 * time.Second).Round(time.Second)) {
		t.Errorf("next != time.Now().Add(1 * time.Second) - %s - %s", next.Round(time.Second), time.Now().Add(1 * time.Second).Round(time.Second))
	}
	if run {
		t.Errorf("Done is Not True")
	}
	next, run = timer.Next()
	if !run {
		t.Errorf("Done is Not True after second run")
	}

	timer.Reschedule(2 * time.Second)

	next, run = timer.Next()
	if !next.Round(time.Second).Equal(time.Now().Add(2 * time.Second).Round(time.Second)) {
		t.Errorf("Reschedule next != time.Now().Add(2 * time.Second) - %s - %s", next.Round(time.Second), time.Now().Add(2 * time.Second).Round(time.Second))
	}

	if run {
		t.Errorf("Done is Not True")
	}
	next, run = timer.Next()
	if !run {
		t.Errorf("Done is Not True after second run")
	}

}

func TestTimerOnceInvalidDuration(t *testing.T) {
	_, err := NewOnce(-1 * time.Second)
	if err == nil {
		t.Errorf("NewOnce Timer Did Not Returned Error")
	}
}

func TestTimerOnceTime(t *testing.T) {
	timer, err := NewOnceTime(time.Now().Add(1 * time.Second))
	if err != nil {
		t.Errorf("NewOnceTime Timer Returned Error %s", err.Error())
	}
	next, run := timer.Next()
	if !next.Round(time.Second).Equal(time.Now().Add(1 * time.Second).Round(time.Second)) {
		t.Errorf("NewOnceTime next != time.Now().Add(1 * time.Second) - %s - %s", next.Round(time.Second), time.Now().Add(1 * time.Second).Round(time.Second))
	}
	if run {
		t.Errorf("NewOnceTime Done is Not True")
	}
	next, run = timer.Next()
	if !run {
		t.Errorf("NewOnceTime Done is Not True after second run")
	}	
}

func TestTimerOnceTimeInvalidDuration(t *testing.T) {
	timer, err := NewOnceTime(time.Now().Add(-1 * time.Hour))
	if err != nil {
		t.Errorf("NewOnce Timer Returned Error %s", err.Error())
	}
	next, run := timer.Next()
	t.Logf("test %d", timer.delay)
	if !run {
		t.Errorf("NewOnceTime Done is True")
	}
	if !next.Round(time.Second).Equal(time.Time{}.Round(time.Second)) {
		t.Errorf("NewOnceTime next != invalid Time - %s - %s", next.Round(time.Second), time.Now().Add(1 * time.Second).Round(time.Second))
	}

}

func TestTimerFixed(t *testing.T) {
	timer, err := NewFixed(10 * time.Second)
	if err != nil {
		t.Errorf("NewFixed Timer Returned Error: %s", err.Error())
	}
	next, run := timer.Next()
	if !next.Round(time.Second).Equal(time.Now().Add(10 * time.Second).Round(time.Second)) {
		t.Errorf("FixedTimer next != time.Now().Add(10 *time.Second) - %s - %s", next.Round(time.Second), time.Now().Add(10 * time.Second).Round(time.Second))
	}
	if run {
		t.Errorf("FixedTimer Run is False")
	}
	timer.Reschedule(2 * time.Second)
	next, run = timer.Next()
	if !next.Round(time.Second).Equal(time.Now().Add(2 * time.Second).Round(time.Second)) {
		t.Errorf("FixedTimer next != time.Now().Add(2 *time.Second) - %s - %s", next.Round(time.Second), time.Now().Add(10 * time.Second).Round(time.Second))
	}
	if run {
		t.Errorf("FixedTimer Run is False")
	}
}

func TestTimerFixedInvalidDuration(t *testing.T) {
	_, err := NewFixed(-1 * time.Second)
	if err == nil {
		t.Errorf("NewOnce Timer Did Not Returned Error")
	}
}

func TestTimerCron(t *testing.T) {
	timer, err := NewCron("5 4 1 12 2")
	if err != nil {
		t.Errorf("Crontimer Timer Returned Error: %s", err.Error())
	}
	next, run := timer.Next()
	test, _ := time.Parse(time.RFC3339, "2021-12-01T04:05:00+08:00")
	if !next.Round(time.Second).Equal(test.Round(time.Second)) {
		t.Errorf("Crontimer next != 2021-12-01T04:05:00+08:00 - %s - %s", next.Round(time.Second), time.Now().Add(10 * time.Second).Round(time.Second))
	}
	if run {
		t.Errorf("Crontimer Run is False")
	}
	timer.Reschedule(10 * time.Second)
	next, run = timer.Next()
	if !next.Round(time.Second).Equal(time.Now().Add(10 * time.Second).Round(time.Second)) {
		t.Errorf("Crontimer next != time.Now().Add(10 *time.Second) - %s - %s", next.Round(time.Second), time.Now().Add(10 * time.Second).Round(time.Second))
	}
	if run {
		t.Errorf("Crontimer Run is False")
	}
}

func TestTimerCronInvalidFormat(t *testing.T) {
	_, err := NewCron("5 4 1 14 2")
	if err == nil {
		t.Errorf("NewOnce Timer Did Not Returned Error")
	}
}