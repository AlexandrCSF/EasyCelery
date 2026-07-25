package worker

import (
	"context"
	"easycelery/internal/executor"
	"easycelery/internal/pipeline"
	"easycelery/internal/pipeline/policy"
	"easycelery/internal/queue"
	"errors"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type WorkerStatuses string

const (
	StatusIdle       WorkerStatuses = "idle"
	StatusProcessing WorkerStatuses = "processing"
)

type Worker struct {
	mu          sync.Mutex
	id          string
	status      WorkerStatuses
	queue       *queue.DefaultQueue
	pipeline    *pipeline.Pipeline
	retryPolicy policy.RetryPolicy
}

func NewWorker(queue *queue.DefaultQueue) *Worker {
	exec := executor.GetDefaultExecutor()

	pipe := pipeline.NewPipeline(pipeline.ExecutionHandler(exec), pipeline.LoggingMiddleware())

	return &Worker{
		id:          uuid.NewString(),
		status:      StatusIdle,
		queue:       queue,
		pipeline:    pipe,
		retryPolicy: policy.NewDefaultRetryPolicy(queue),
	}
}
func (w *Worker) ID() string {
	return w.id
}
func (w *Worker) Status() WorkerStatuses {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}

func (w *Worker) SetStatus(status WorkerStatuses) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = status
}

func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-w.queue.NotificationChannel():
			processedTask, err := w.queue.Pop()
			if err != nil {
				if errors.Is(err, queue.ErrEmpty) {
					break
				}
				slog.Error("Got an error while processing task",
					"worker ID: ", w.id,
					"error: ", err)
				continue
			}
			w.SetStatus(StatusProcessing)

			err = w.pipeline.Execute(
				ctx,
				processedTask,
			)

			w.SetStatus(StatusIdle)

			w.retryPolicy.Handle(processedTask, err)

			if err == nil {
				w.queue.Notify()
			}
		case <-ctx.Done():
			slog.Error("Worker stopping due to context stop", "id:", w.id)
			return
		}
	}
}
