package worker

import (
	"context"
	"easycelery/internal/executor"
	"easycelery/internal/queue"
	"easycelery/internal/task"
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
	mu     sync.Mutex
	id     string
	status WorkerStatuses
	queue  *queue.DefaultQueue
}

func NewWorker(queue *queue.DefaultQueue) *Worker {
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

func (w *Worker) TakeOnATask(task *task.Task, ctx context.Context) error {
	w.SetStatus(StatusProcessing)
	defer w.SetStatus(StatusIdle)

	res, err := executor.GetDefaultExecutor().Process(task, ctx)
	if err != nil {
		return err
	}
	slog.Info("Task completed",
		"task_id", task.ID(),
		"result", res,
		"worker", w.id,
	)
	w.queue.Notify()
	return nil
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
			err = w.TakeOnATask(processedTask, ctx)
			if err != nil {
				slog.Error("Got an error while processing task",
					"worker ID", w.id,
					"error", err,
					"attempting retry in", processedTask.RetryDelay())
				delay, ok := processedTask.TryScheduleRetry()
				if ok && delay != nil {
					w.queue.PushLater(ctx, processedTask, *delay)
				} else {
					slog.Error("Cannot send task to retry!",
						"task_id", processedTask.ID())
				}
			}
		case <-ctx.Done():
			slog.Error("Worker stopping due to context stop", "id:", w.id)
			return
		}
	}
}
