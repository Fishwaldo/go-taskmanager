package taskmanager

import (
	"sync"

	"github.com/Fishwaldo/go-taskmanager/job"
)

type jobMap struct {
	jobs map[string]*job.Job
	mx   sync.RWMutex
}

func newJobMap() *jobMap {
	return &jobMap{
		jobs: make(map[string]*job.Job),
	}
}

func (jm *jobMap) add(j *job.Job) {
	jm.mx.Lock()
	defer jm.mx.Unlock()
	jm.jobs[j.ID()] = j
}

func (jm *jobMap) delete(j *job.Job) {
	jm.mx.Lock()
	defer jm.mx.Unlock()
	delete(jm.jobs, j.ID())
}

func (jm *jobMap) len() int {
	jm.mx.RLock()
	defer jm.mx.RUnlock()
	return len(jm.jobs)
}
