package executor

import (
	"context"
	"easycelery/internal/task"
	"errors"
	"sync"
)

var (
	once sync.Once
	inst *DefaultExecutor
)

type Executor interface {
	Process(task *task.Task, ctx context.Context) (any, error)
}
type DefaultExecutor struct {
}

func GetDefaultExecutor() *DefaultExecutor {
	once.Do(func() {
		inst = &DefaultExecutor{}
	})
	return inst
}

func (executor *DefaultExecutor) Process(t *task.Task, ctx context.Context) (any, error) {
	t.SetStatus(task.StatusProcessing)
	t.SetStartAt()

	if t.Func() == nil {
		t.SetStatus(task.StatusCompleted)
		t.SetCompletedAtNow()
		return nil, errors.New(t.Error())
	}

	if err := ctx.Err(); err != nil {
		t.SetStatus(task.StatusError)
		t.SetError(err.Error())
		t.SetCompletedAtNow()
		return nil, err
	}

	res, err := t.Func()(ctx)
	if err != nil {
		t.SetStatus(task.StatusError)
		t.SetError(err.Error())
		t.SetCompletedAtNow()
		return nil, err
	}

	t.SetStatus(task.StatusCompleted)
	t.SetResult(res)
	t.SetCompletedAtNow()
	return res, nil
}
