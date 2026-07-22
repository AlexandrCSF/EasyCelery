package task

import (
	"context"
	"log/slog"
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
	TryScheduleRetry(delay time.Duration) bool
}
type Task struct {
	mu sync.RWMutex
	id string

	createdAt time.Time
	startAt   time.Time

	completedAt time.Time
	result      any
	error       string

	status TaskStatuses

	fn TaskFunc

	retryCount int
	maxRetries int
	retryDelay time.Duration
}

func NewTask(fn TaskFunc) *Task {
	newId := uuid.NewString()
	slog.Info("New task passed",
		"id:", newId)
	return &Task{
		id:        newId,
		status:    StatusPlan,
		createdAt: time.Now(),
		fn:        fn,
	}
}

func (t *Task) Status() TaskStatuses {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *Task) CompletedAt() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.completedAt
}
func (t *Task) StartAt() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.startAt
}
func (t *Task) Error() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.error
}

func (t *Task) SetError(error string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.error = error
}
func (t *Task) SetCompletedAt(time time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.completedAt = time
}

func (t *Task) SetCompletedAtNow() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.completedAt = time.Now()
}
func (t *Task) SetStatus(status TaskStatuses) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = status
}

func (t *Task) SetResult(result any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.result = result
}
func (t *Task) GetResult() any {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.result
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
	t.startAt = time.Now()
}

func (t *Task) TryScheduleRetry(delay time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.retryCount >= t.maxRetries {
		return false
	}
	t.retryCount++
	t.retryDelay = delay
	return true
}
