package task

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TaskStatuses string

const (
	StatusPlan       TaskStatuses = "plan"
	StatusCompleted  TaskStatuses = "completed"
	StatusProcessing TaskStatuses = "processing"
	StatusError      TaskStatuses = "error"
)

type TaskFunc func() (any, error)

type Task struct {
	id string

	fn         TaskFunc
	parameters []string

	createdAt time.Time
	startAt   time.Time

	completedAt time.Time
	result      any
	error       string

	status TaskStatuses
}

func NewTask(fn TaskFunc, params []string) *Task {
	return &Task{
		id:         uuid.NewString(),
		status:     StatusPlan,
		fn:         fn,
		parameters: params,
		createdAt:  time.Now(),
	}
}
func (task *Task) Id() string {
	return task.id
}

func (t *Task) Process() (any, error) {
	t.status = StatusProcessing
	t.startAt = time.Now()

	if t.fn == nil {
		t.status = StatusCompleted
		t.completedAt = time.Now()
		return nil, errors.New(t.error)
	}

	res, err := t.fn()
	if err != nil {
		t.status = StatusError
		t.error = err.Error()
		t.completedAt = time.Now()
		return nil, err
	}

	t.status = StatusCompleted
	t.result = res
	t.completedAt = time.Now()
	return res, nil
}
