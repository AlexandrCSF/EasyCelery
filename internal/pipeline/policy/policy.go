package policy

import (
	"context"
	"easycelery/internal/queue"
	"easycelery/internal/task"
)

type RetryPolicy interface {
	Handle(ctx context.Context, task *task.Task, err error)
}

type DefaultRetryPolicy struct {
	queue queue.Queue
}

func NewDefaultRetryPolicy(q queue.Queue) *DefaultRetryPolicy {
	return &DefaultRetryPolicy{
		queue: q,
	}
}

func (p *DefaultRetryPolicy) Handle(
	ctx context.Context,
	task *task.Task,
	err error,
) {

	if err == nil {
		return
	}

	delay, ok := task.TryScheduleRetry()
	if !ok {
		return
	}

	p.queue.PushLater(task, *delay)
}
