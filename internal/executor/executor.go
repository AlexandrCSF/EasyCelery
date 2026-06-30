package executor

import (
	"easycelery/internal/task"
	"errors"
	"sync"
	"time"
)

var (
	once sync.Once
	inst *DefaultExecutor
)

type Executor interface {
	Process(task *task.Task) (any, error)
	Auth() (any, error)
}
type DefaultExecutor struct {
}

func GetDefaultExecutor() *DefaultExecutor {
	once.Do(func() {
		inst = &DefaultExecutor{}
	})
	return inst
}
func (e *DefaultExecutor) Auth() (any, error) {
	return nil, nil
}
func (executor *DefaultExecutor) Process(t *task.Task) (any, error) {
	t.SetStatus(task.StatusProcessing)
	t.SetStartAt()

	if t.Func() == nil {
		t.SetStatus(task.StatusCompleted)
		t.CompletedAt = time.Now()
		return nil, errors.New(t.Error)
	}

	res, err := t.Func()()
	if err != nil {
		t.SetStatus(task.StatusError)
		t.Error = err.Error()
		t.CompletedAt = time.Now()
		return nil, err
	}

	t.SetStatus(task.StatusCompleted)
	t.Result = res
	t.CompletedAt = time.Now()
	return res, nil
}
