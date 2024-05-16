package sdk

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
)

type Job interface {
	Id() string
	Description() string
	Run() error
}

type JobQueue struct {
	queue         chan Job
	maxConcurrent int
	stream        golang.Plugin_RegisterClient
}

func NewJobQueue(maxConcurrent int, stream golang.Plugin_RegisterClient) *JobQueue {
	return &JobQueue{
		queue:         make(chan Job, 10000),
		maxConcurrent: maxConcurrent,
		stream:        stream,
	}
}

func (q *JobQueue) Push(job Job) {
	q.stream.Send(&golang.PluginMessage{
		PluginMessage: &golang.PluginMessage_Job{
			Job: &golang.JobResult{
				Id:             job.Id(),
				Description:    job.Description(),
				FailureMessage: "",
				Done:           false,
			},
		},
	})

	q.queue <- job
}

func (q *JobQueue) Start() {
	for i := 0; i < q.maxConcurrent; i++ {
		go q.run()
	}
}

func (q *JobQueue) run() {
	defer func() {
		if r := recover(); r != nil {
			q.stream.Send(&golang.PluginMessage{
				PluginMessage: &golang.PluginMessage_Err{
					Err: &golang.Error{Error: fmt.Sprintf("%v", r)},
				},
			})

			go q.run()
		}
	}()

	for job := range q.queue {
		jobResult := &golang.JobResult{
			Id:          job.Id(),
			Description: job.Description(),
			Done:        true,
		}
		if err := job.Run(); err != nil {
			jobResult.FailureMessage = err.Error()
		}

		q.stream.Send(&golang.PluginMessage{
			PluginMessage: &golang.PluginMessage_Job{
				Job: jobResult,
			},
		})
	}
}
