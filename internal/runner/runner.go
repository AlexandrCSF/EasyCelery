package runner

import (
	"context"
	"easycelery/internal/queue"
	"easycelery/internal/task"
	"log/slog"
)

const defaultConcurrency = 1

type Runner interface {
	Run()
	SendTask(task task.Task)
}

type DefaultRunner struct {
	queue       queue.Queue
	concurrency int
}

func NewDefaultRunner(q queue.Queue, concurrency int) *DefaultRunner {
	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}
	return &DefaultRunner{queue: q, concurrency: concurrency}
}

func NewDefaultRunnerDefaultValues(q queue.Queue) *DefaultRunner {
	return &DefaultRunner{queue: q, concurrency: defaultConcurrency}
}

func (r *DefaultRunner) SendTask(t task.Task) {
	r.queue.Push(t)
}

func (r *DefaultRunner) Run(ctx context.Context) {
	for {
		select {
		case <-r.queue.GetNotificationChannel():
			if err := r.queue.HandleNext(ctx); err != nil {
				slog.Error("error processing next task", "error", err)
			}
		case <-ctx.Done():
			return
		default:
		}
	}
}
