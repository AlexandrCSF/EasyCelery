package queue

import (
	"easycelery/internal/task"
	"errors"
	"sync"
	"time"
)

var ErrEmpty = errors.New("queue is empty")

type Queue interface {
	Push(task *task.Task)
	Pop() (*task.Task, error)
	Length() int
	PushLater(t *task.Task, delay time.Duration)
}

type DefaultQueue struct {
	mu                  sync.RWMutex
	tasksToProcess      []*task.Task
	completedTasks      []*task.Task
	notificationChannel chan struct{}
	maxWorkers          int
	scheduler           *Scheduler
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
	queue := &DefaultQueue{
		notificationChannel: make(chan struct{}, workers),
		maxWorkers:          workers,
	}
	queue.scheduler = NewScheduler(queue)
	return queue
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

func (q *DefaultQueue) PushLater(t *task.Task, delay time.Duration) {
	q.scheduler.Add(t, delay)
}

func (q *DefaultQueue) Scheduler() *Scheduler {
	return q.scheduler
}
