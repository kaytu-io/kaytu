package controller

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"sort"
	"sync"
	"time"
)

type Jobs struct {
	runningJobsMap map[string]string
	failedJobsMap  map[string]string

	statusErr string

	jobMutex  sync.RWMutex
	jobChan   chan *golang.JobResult
	errorChan chan error
}

func NewJobs() *Jobs {
	jobs := Jobs{
		runningJobsMap: map[string]string{},
		failedJobsMap:  map[string]string{},
		statusErr:      "",
		jobMutex:       sync.RWMutex{},
		jobChan:        make(chan *golang.JobResult, 10000),
		errorChan:      make(chan error, 10000),
	}
	go jobs.UpdateStatus()

	return &jobs
}

func (m *Jobs) RunningJobs() []string {
	m.jobMutex.RLock()
	defer m.jobMutex.RUnlock()

	if len(m.runningJobsMap) == 0 {
		return nil
	}
	var res []string
	for _, v := range m.runningJobsMap {
		res = append(res, v)
	}
	sort.Strings(res)
	return res
}

func (m *Jobs) RunningJobsSummary() ([]string, bool) {
	m.jobMutex.RLock()
	defer m.jobMutex.RUnlock()

	if len(m.runningJobsMap) == 0 {
		return nil, false
	}
	var res []string
	for _, v := range m.runningJobsMap {
		res = append(res, v)
	}
	sort.Strings(res)
	count := 3
	if len(res) < 3 {
		count = len(res)
	}
	return res[:count], len(m.runningJobsMap) > 3
}

func (m *Jobs) FailedJobs() []string {
	m.jobMutex.RLock()
	defer m.jobMutex.RUnlock()

	if len(m.failedJobsMap) == 0 {
		return nil
	}
	var res []string
	for _, v := range m.failedJobsMap {
		res = append(res, v)
	}
	sort.Strings(res)
	return res
}

func (m *Jobs) FailedJobsSummary() ([]string, bool) {
	m.jobMutex.RLock()
	defer m.jobMutex.RUnlock()

	if len(m.failedJobsMap) == 0 {
		return nil, false
	}
	var res []string
	for _, v := range m.failedJobsMap {
		res = append(res, v)
	}
	sort.Strings(res)
	count := 3
	if len(res) < 3 {
		count = len(res)
	}
	return res[:count], len(m.failedJobsMap) > 3
}

func (m *Jobs) UpdateStatus() {
	for {
		select {
		case job := <-m.jobChan:
			m.jobMutex.Lock()
			if !job.Done {
				m.runningJobsMap[job.Id] = job.Description
			} else {
				if _, ok := m.runningJobsMap[job.Id]; ok {
					delete(m.runningJobsMap, job.Id)
				}
			}
			if len(job.FailureMessage) > 0 {
				m.failedJobsMap[job.Id] = fmt.Sprintf("%s failed due to %s", job.Description, job.FailureMessage)
			}
			m.jobMutex.Unlock()

		case err := <-m.errorChan:
			m.statusErr = fmt.Sprintf("%s\nFailed due to %v", m.statusErr, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *Jobs) PublishError(err error) {
	m.errorChan <- err
}

func (m *Jobs) Publish(job *golang.JobResult) *golang.JobResult {
	m.jobChan <- job
	return job
}

func (m *Jobs) GetError() string {
	return m.statusErr
}
