package runner

import (
	"context"
	"easycelery/internal/queue"
	"easycelery/internal/worker"
	"log/slog"
)

const defaultNumWorkers = 1

type Runner interface {
	Run(ctx context.Context)
}

type DefaultRunner struct {
	queue   *queue.DefaultQueue
	workers []*worker.Worker
}

func NewDefaultRunner(q *queue.DefaultQueue, numWorkers int) *DefaultRunner {
	if numWorkers <= 0 {
		numWorkers = defaultNumWorkers
	}
	var workers []*worker.Worker
	for i := 0; i < numWorkers; i++ {
		workers = append(workers, worker.NewWorker(q))
	}
	return &DefaultRunner{queue: q, workers: workers}
}

func NewDefaultRunnerDefaultValues(q *queue.DefaultQueue) *DefaultRunner {
	var workers []*worker.Worker
	for i := 0; i < defaultNumWorkers; i++ {
		workers = append(workers, worker.NewWorker(q))
	}
	return &DefaultRunner{queue: q, workers: workers}
}

func (r *DefaultRunner) Run(ctx context.Context) {
	for {
		select {
		case <-r.queue.NotificationChannel():
			r.dispatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (r *DefaultRunner) dispatch(ctx context.Context) {
	for r.queue.HasNext() {
		w := r.FindIdleWorker()
		if w == nil {
			return
		}
		go func(w *worker.Worker) {
			if err := w.TakeOnATask(ctx); err != nil {
				slog.Error("error processing task", "error", err, "worker", w.ID())
			}
			r.queue.Notify()
		}(w)
	}
}
func (r *DefaultRunner) FindIdleWorker() *worker.Worker {
	for _, w := range r.workers {
		if w.TryStart() {
			return w
		}
	}
	return nil
}
