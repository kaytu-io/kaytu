package sdk

import (
	"context"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"log"
	"runtime/debug"
	"sync/atomic"
	"time"
)

type Job interface {
	Id() string
	Description() string
	Run(ctx context.Context) error
}

type JobQueue struct {
	queue         chan Job
	maxConcurrent int
	stream        *StreamController

	pendingCounter  atomic.Uint32
	finishedCounter atomic.Uint32
	onFinish        func(ctx context.Context)
}

func NewJobQueue(maxConcurrent int, stream *StreamController) *JobQueue {
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

func (q *JobQueue) finisher(ctx context.Context) {
	if err := recover(); err != nil {
		log.Printf("Job queue finisher panic: %v", err)
		go q.finisher(ctx)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	lastTickLog := time.Now()
	for range ticker.C {
		if time.Since(lastTickLog) > time.Minute {
			log.Printf("Job queue finisher: %d/%d", q.finishedCounter.Load(), q.pendingCounter.Load())
			lastTickLog = time.Now()
		}
		if q.finishedCounter.Load() == q.pendingCounter.Load() && q.onFinish != nil {
			time.Sleep(500 * time.Millisecond)
			log.Printf("All jobs are finished - calling onFinish, job counts: %d/%d", q.finishedCounter.Load(), q.pendingCounter.Load())
			q.onFinish(ctx)
		}
	}
}

func (q *JobQueue) Start(ctx context.Context) {
	for i := 0; i < q.maxConcurrent; i++ {
		go q.run(ctx)
	}
	go q.finisher(ctx)
}

func (q *JobQueue) SetOnFinish(f func(ctx context.Context)) {
	q.onFinish = f
}

func (q *JobQueue) handleJob(ctx context.Context, job Job) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Job queue handle job panic: %v, stack: %v", r, debug.Stack())
			q.stream.Send(&golang.PluginMessage{
				PluginMessage: &golang.PluginMessage_Err{
					Err: &golang.Error{Error: fmt.Sprintf("job %s paniced: %v", job.Id(), r)},
				},
			})
		}
	}()
	defer q.finishedCounter.Add(1)

	jobResult := &golang.JobResult{
		Id:          job.Id(),
		Description: job.Description(),
		Done:        true,
	}
	log.Printf("Running job %s", job.Id())
	if err := job.Run(ctx); err != nil {
		jobResult.FailureMessage = err.Error()
		log.Printf("Failed job %s: %s", job.Id(), err.Error())
	} else {
		log.Printf("Finished job %s", job.Id())
	}

	q.stream.Send(&golang.PluginMessage{
		PluginMessage: &golang.PluginMessage_Job{
			Job: jobResult,
		},
	})
}

func (q *JobQueue) run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Job queue run panic: %v", r)
			q.stream.Send(&golang.PluginMessage{
				PluginMessage: &golang.PluginMessage_Err{
					Err: &golang.Error{Error: fmt.Sprintf("%v", r)},
				},
			})

			go q.run(ctx)
		}
	}()

	for job := range q.queue {
		q.handleJob(ctx, job)
	}
}
