package runner

import (
	"context"
	"easycelery/internal/queue"
	"easycelery/internal/worker"
	"log/slog"
	"sync"
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
	var wg sync.WaitGroup
	for _, w := range r.workers {
		wg.Add(1)
		go func(w *worker.Worker) {
			defer wg.Done()
			w.Run(ctx)
		}(w)
	}

	<-ctx.Done()
	slog.Info("Runner stopping due to context stop")
	wg.Wait()
}
