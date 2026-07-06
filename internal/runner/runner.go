package runner

import (
	"context"
	"easycelery/internal/queue"
	"easycelery/internal/worker"
	"log/slog"
)

const defaultNumWorkers = 1

type Runner interface {
	Config
	Run(ctx context.Context)
}
type Config struct {
	Workers int
}
type DefaultRunner struct {
	queue   *queue.DefaultQueue
	workers []*worker.Worker
}

func NewDefaultRunner(q *queue.DefaultQueue, config Config) *DefaultRunner {
	if config.Workers <= 0 {
		slog.Error("numWorkers parameter must be bigger than 0")
		return nil
	}
	var workers []*worker.Worker
	for i := 0; i < config.Workers; i++ {
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
			slog.Error("Runner stopping due to context stop")
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
