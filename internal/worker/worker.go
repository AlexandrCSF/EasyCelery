package worker

import (
	"context"
	"easycelery/internal/executor"
	"easycelery/internal/queue"
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
	mu     sync.Mutex
	id     string
	status WorkerStatuses
	queue  queue.Queue
}

func NewWorker(queue queue.Queue) *Worker {
	return &Worker{
		id:     uuid.NewString(),
		status: StatusIdle,
		queue:  queue,
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

func (w *Worker) TakeOnATask(ctx context.Context) error {
	defer w.SetStatus(StatusIdle)

	task, err := w.queue.Pop()
	if err != nil {
		slog.Error("Worker errored while trying to get task from queue", "worker", w.id, "error", err)
		return err
	}
	res, err := executor.GetDefaultExecutor().Process(task, ctx)
	if err != nil {
		slog.Error("Got an error while processing task",
			"task_id", task.ID(),
			"error", err,
			"worker", w.id)
		return err
	}
	slog.Info("Task completed",
		"task_id", task.ID(),
		"result", res,
		"worker", w.id,
	)
	return nil
}

func (w *Worker) TryStart() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.status != StatusIdle {
		return false
	}
	w.status = StatusProcessing
	return true
}
