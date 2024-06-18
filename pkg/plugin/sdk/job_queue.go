package sdk

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"log"
	"sync/atomic"
	"time"
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

	pendingCounter  atomic.Uint32
	finishedCounter atomic.Uint32
	onFinish        func()
}

func NewJobQueue(maxConcurrent int, stream golang.Plugin_RegisterClient) *JobQueue {
	return &JobQueue{
		queue:         make(chan Job, 10000),
		maxConcurrent: maxConcurrent,
		stream:        stream,

		pendingCounter:  atomic.Uint32{},
		finishedCounter: atomic.Uint32{},
	}
}

func (q *JobQueue) Push(job Job) {
	log.Printf("Pushing job %s to queue", job.Id())
	q.pendingCounter.Add(1)

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

	go func() {
		for {
			if q.finishedCounter.Load() == q.pendingCounter.Load() && q.onFinish != nil {
				time.Sleep(500 * time.Millisecond)
				log.Printf("All jobs are finished - calling onFinish, job counts: %d/%d", q.finishedCounter.Load(), q.pendingCounter.Load())
				q.onFinish()
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func (q *JobQueue) SetOnFinish(f func()) {
	q.onFinish = f
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
		log.Printf("Running job %s", job.Id())
		if err := job.Run(); err != nil {
			jobResult.FailureMessage = err.Error()
		}
		log.Printf("Finished job %s", job.Id())

		q.stream.Send(&golang.PluginMessage{
			PluginMessage: &golang.PluginMessage_Job{
				Job: jobResult,
			},
		})
		q.finishedCounter.Add(1)
	}
}
