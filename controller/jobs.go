package controller

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"sort"
	"sync"
	"time"
)

type Jobs struct {
	runningJobsMap sync.Map
	failedJobsMap  sync.Map

	statusErr string

	jobChan   chan *golang.JobResult
	errorChan chan error
}

func NewJobs() *Jobs {
	jobs := Jobs{
		runningJobsMap: sync.Map{},
		failedJobsMap:  sync.Map{},
		statusErr:      "",
		jobChan:        make(chan *golang.JobResult, 10000),
		errorChan:      make(chan error, 10000),
	}
	go jobs.UpdateStatus()

	return &jobs
}

func (m *Jobs) RunningJobs() []string {
	var res []string
	m.runningJobsMap.Range(func(key, value any) bool {
		res = append(res, value.(string))
		return true
	})
	sort.Strings(res)
	return res
}

func (m *Jobs) FailedJobs() []string {
	var res []string
	m.failedJobsMap.Range(func(key, value any) bool {
		res = append(res, value.(string))
		return true
	})
	sort.Strings(res)
	return res
}

func (m *Jobs) UpdateStatus() {
	for {
		select {
		case job := <-m.jobChan:
			if !job.Done {
				m.runningJobsMap.Store(job.Id, job.Description)
			} else {
				m.runningJobsMap.Delete(job.Id)
			}
			if len(job.FailureMessage) > 0 {
				m.failedJobsMap.Store(job.Id, fmt.Sprintf("%s failed due to %s", job.Description, job.FailureMessage))
			}
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
