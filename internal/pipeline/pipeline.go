package pipeline

import (
	"context"
	"easycelery/internal/queue"
	"easycelery/internal/task"
)

type ExecutionContext struct {
	Context context.Context

	Task *task.Task

	Queue queue.Queue

	Result any

	Error error
}

type Pipeline struct {
	handler Handler
}

func NewPipeline(
	handler Handler,
	middlewares ...Middleware,
) *Pipeline {

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return &Pipeline{
		handler: handler,
	}
}
func (p *Pipeline) Execute(
	ctx context.Context,
	task *task.Task,
) error {

	return p.handler(ctx, task)
}
