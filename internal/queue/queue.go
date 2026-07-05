package queue

import (
	"context"
	"easycelery/internal/executor"
	"easycelery/internal/task"
	"errors"
	"log/slog"
	"sync"
)

type Queue interface {
	Push(task *task.Task)
	Pop() (*task.Task, error)
}

type DefaultQueue struct {
	mu                  sync.RWMutex
	tasksToProcess      []*task.Task
	completedTasks      []*task.Task
	erroredTasks        []*task.Task
	notificationChannel chan struct{}
}

func (q *DefaultQueue) NotificationChannel() chan struct{} {
	return q.notificationChannel
}

func NewInMemoryQueue() *DefaultQueue {
	return &DefaultQueue{
		notificationChannel: make(chan struct{}, 1),
	}
}

func (q *DefaultQueue) Notify() {
	q.NotificationChannel() <- struct{}{}
}

func (q *DefaultQueue) Push(task *task.Task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasksToProcess = append(q.tasksToProcess, task)
	q.Notify()
}

func (q *DefaultQueue) HasNext() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tasksToProcess) > 0
}

func (q *DefaultQueue) Pop() (*task.Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tasksToProcess) == 0 {
		return nil, errors.New("queue is empty")
	}
	poppedTask := q.tasksToProcess[0]
	q.tasksToProcess = q.tasksToProcess[1:]
	return poppedTask, nil
}

func (q *DefaultQueue) HandleNext(ctx context.Context) error {
	taskToProcess, err := q.Pop()
	if err != nil {
		if taskToProcess != nil {
			q.erroredTasks = append(q.erroredTasks, taskToProcess)
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
		q.erroredTasks = append(q.erroredTasks, taskToProcess)
		return err
	}
	q.erroredTasks = append(q.completedTasks, taskToProcess)
	return nil
}
