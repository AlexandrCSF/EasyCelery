package queue

import (
	"context"
	"easycelery/internal/executor"
	"easycelery/internal/task"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var ErrEmpty = errors.New("queue is empty")

type Queue interface {
	Push(task *task.Task)
	Pop() (*task.Task, error)
	Length() int
	PushLater(ctx context.Context, t *task.Task, delay time.Duration)
}

type DefaultQueue struct {
	mu                  sync.RWMutex
	tasksToProcess      []*task.Task
	completedTasks      []*task.Task
	notificationChannel chan struct{}
	maxWorkers          int
}

func (q *DefaultQueue) NotificationChannel() chan struct{} {
	return q.notificationChannel
}
func (q *DefaultQueue) Length() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tasksToProcess)
}
func NewInMemoryQueue(workers int) *DefaultQueue {
	return &DefaultQueue{
		notificationChannel: make(chan struct{}, workers),
		maxWorkers:          workers,
	}
}

func (q *DefaultQueue) Notify() {
	q.NotifyN(1)
}

func (q *DefaultQueue) NotifyN(n int) {
	if n > q.maxWorkers {
		n = q.maxWorkers
	}
	for i := 0; i < n; i++ {
		select {
		case q.notificationChannel <- struct{}{}:
		default:
			return
		}
	}
}

func (q *DefaultQueue) Push(task *task.Task) {
	q.mu.Lock()
	q.tasksToProcess = append(q.tasksToProcess, task)
	pending := len(q.tasksToProcess)
	q.mu.Unlock()
	q.NotifyN(pending)
}

func (q *DefaultQueue) Pop() (*task.Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tasksToProcess) == 0 {
		return nil, ErrEmpty
	}
	poppedTask := q.tasksToProcess[0]
	q.tasksToProcess = q.tasksToProcess[1:]
	return poppedTask, nil
}

func (q *DefaultQueue) HandleNext(ctx context.Context) error {
	taskToProcess, err := q.Pop()
	if err != nil {
		if taskToProcess != nil {
			return err
		} else {
			return err
		}
	}
	res, err := executor.GetDefaultExecutor().Process(taskToProcess, ctx)
	slog.Info("Task completed",
		"task_id", taskToProcess.ID(),
		"result", res,
	)
	if err != nil {
		return err
	}
	q.completedTasks = append(q.completedTasks, taskToProcess)
	return nil
}

func (q *DefaultQueue) PushLater(ctx context.Context, t *task.Task, delay time.Duration) {
	go func() {
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
			q.Push(t)
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}()
}
