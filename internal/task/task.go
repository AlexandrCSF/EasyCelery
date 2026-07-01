package task

import (
	"context"
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

type TaskFunc func(ctx context.Context) (any, error)
type Task struct {
	id string

	CreatedAt time.Time
	StartAt   time.Time

	CompletedAt time.Time
	Result      any
	Error       string

	status TaskStatuses

	fn TaskFunc
}

func NewTask(fn TaskFunc) *Task {
	return &Task{
		id:        uuid.NewString(),
		status:    StatusPlan,
		CreatedAt: time.Now(),
		fn:        fn,
	}
}

func (t *Task) Status() TaskStatuses {
	return t.status
}
func (t *Task) SetStatus(status TaskStatuses) {
	t.status = status
}

func (t *Task) Func() TaskFunc {
	return t.fn
}
func (t *Task) ID() string {
	return t.id
}

func (t *Task) SetStartAt() {
	t.StartAt = time.Now()
}
