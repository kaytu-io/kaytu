package sdk

import (
	"context"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"log"
	"runtime/debug"
	"sync/atomic"
	"time"
)

type JobProperties struct {
	ID          string
	Description string
	MaxRetry    int
}

type Job interface {
	Properties() JobProperties
	Run(ctx context.Context) error
}

type JobQueue struct {
	queue         chan Job
	maxConcurrent int
	stream        *StreamController

	pendingCounter  atomic.Uint32
	finishedCounter atomic.Uint32
	onFinish        func(ctx context.Context)
	retryCount      utils.ConcurrentMap[string, int]
}

func NewJobQueue(maxConcurrent int, stream *StreamController) *JobQueue {
	return &JobQueue{
		queue:         make(chan Job, 10000),
		maxConcurrent: maxConcurrent,
		stream:        stream,
		retryCount:    utils.NewConcurrentMap[string, int](),

		pendingCounter:  atomic.Uint32{},
		finishedCounter: atomic.Uint32{},
	}
}

func (q *JobQueue) Push(job Job) {
	props := job.Properties()
	log.Printf("Pushing job %s to queue", props.ID)
	q.pendingCounter.Add(1)

	q.stream.Send(&golang.PluginMessage{
		PluginMessage: &golang.PluginMessage_Job{
			Job: &golang.JobResult{
				Id:             props.ID,
				Description:    props.Description,
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
		if err := ctx.Err(); err != nil {
			log.Printf("Job queue finisher context error: %v", err)
			return
		}

		if time.Since(lastTickLog) > time.Minute {
			log.Printf("Job queue finisher: %d/%d", q.finishedCounter.Load(), q.pendingCounter.Load())
			lastTickLog = time.Now()
		}
		if q.finishedCounter.Load() == q.pendingCounter.Load() && q.onFinish != nil {
			time.Sleep(5000 * time.Millisecond)
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

func (q *JobQueue) runJob(ctx context.Context, job Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("job paniced: %v, stack: %v", r, string(debug.Stack()))
		}
	}()

	return job.Run(ctx)
}

func (q *JobQueue) handleJob(ctx context.Context, job Job) {
	props := job.Properties()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Job queue handle job panic: %v, stack: %v", r, string(debug.Stack()))
			q.stream.Send(&golang.PluginMessage{
				PluginMessage: &golang.PluginMessage_Err{
					Err: &golang.Error{Error: fmt.Sprintf("job %s paniced: %v", props.ID, r)},
				},
			})
		}
	}()
	defer q.finishedCounter.Add(1)

	jobResult := &golang.JobResult{
		Id:          props.ID,
		Description: props.Description,
		Done:        true,
	}
	log.Printf("Running job %s", props.ID)
	if err := q.runJob(ctx, job); err != nil {
		jobResult.FailureMessage = err.Error()
		if v, ok := q.retryCount.Get(props.ID); v < props.MaxRetry {
			if !ok {
				v2, loaded := q.retryCount.LoadOrStore(props.ID, 0)
				if loaded {
					v = v2
				}
			}
			for !q.retryCount.CompareAndSwap(props.ID, v, v+1) {
				v, _ = q.retryCount.Get(props.ID)
			}
			if v+1 < props.MaxRetry {
				log.Printf("Failed job %s: %s, retrying[%d/%d]", props.ID, err.Error(), v+1, props.MaxRetry)
				q.Push(job)
				return
			} else {
				log.Printf("Failed job %s: %s", props.ID, err.Error())
			}
		} else {
			log.Printf("Failed job %s: %s", props.ID, err.Error())
		}
	} else {
		log.Printf("Finished job %s", props.ID)
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
		if err := ctx.Err(); err != nil {
			log.Printf("context error: %v", err)
			return
		}
		q.handleJob(ctx, job)
	}
}
