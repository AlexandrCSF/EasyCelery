package pipeline

import (
	"context"
	"easycelery/internal/executor"
	"easycelery/internal/task"
)

type Handler func(ctx context.Context, task *task.Task) error

func ExecutionHandler(exec executor.Executor) Handler {
	return func(ctx context.Context, t *task.Task) error {
		_, err := exec.Process(t, ctx)
		return err
	}
}
