package task

import (
	"context"
	"sync"
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

type BaseTask interface {
	Status() TaskStatuses
	SetStatus(status TaskStatuses)
	ID() string
	SetStartAt()
}
type Task struct {
	mu sync.RWMutex
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
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}
func (t *Task) SetStatus(status TaskStatuses) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = status
}

func (t *Task) Func() TaskFunc {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.fn
}
func (t *Task) ID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.id
}

func (t *Task) SetStartAt() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.StartAt = time.Now()
}
