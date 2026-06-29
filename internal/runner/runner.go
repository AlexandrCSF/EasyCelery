package runner

import (
	"easycelery/internal/queue"
	"easycelery/internal/task"
	"log/slog"
)

const defaultConcurrency = 1

type Runner interface {
	RunExecutionForever()
	SendTask(task *task.Task)
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

func (r *DefaultRunner) RunExecutionForever() {
	for {
		if r.queue.HasNext() {
			err := r.queue.ProcessNext()
			if err != nil {
				slog.Error("Error processing next task: %s", err)
			}
		}
	}
}
