package runner

import "easycelery/internal/queue"

const defaultConcurrency = 1

type Runner interface {
	RunExecutionForever()
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

func (r *DefaultRunner) RunExecutionForever() {
	for {
		err := r.queue.ProcessNext()
	}
}
