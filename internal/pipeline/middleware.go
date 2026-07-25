package pipeline

import (
	"context"
	"easycelery/internal/task"
	"log/slog"
	"time"
)

type Middleware func(Handler) Handler

func Chain(handler Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func LoggingMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, task *task.Task) error {
			start := time.Now()

			slog.Info(
				"task started",
				"task_id",
				task.ID(),
			)

			err := next(ctx, task)

			slog.Info(
				"task finished",
				"task_id",
				task.ID(),
				"duration",
				time.Since(start),
				"error",
				err,
			)
			return err
		}
	}
}
