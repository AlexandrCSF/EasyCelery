package queue

import (
	"context"
	"easycelery/internal/executor"
	"easycelery/internal/task"
	"errors"
	"log/slog"
)

type Queue interface {
	Push(task task.Task)
	Pop() (*task.Task, error)
	ProcessNext(ctx context.Context) error
	HasNext() bool
}

type InMemoryQueue struct {
	tasksToProcess []task.Task
	completedTasks []task.Task
	erroredTasks   []task.Task
}

func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{}
}

func (q *InMemoryQueue) Push(task task.Task) {
	q.tasksToProcess = append(q.tasksToProcess, task)
}
func (q *InMemoryQueue) HasNext() bool {
	return len(q.tasksToProcess) > 0
}

func (q *InMemoryQueue) Pop() (*task.Task, error) {
	if len(q.tasksToProcess) == 0 {
		return nil, errors.New("queue is empty")
	}
	poppedTask := q.tasksToProcess[0]
	q.tasksToProcess = q.tasksToProcess[1:]
	return &poppedTask, nil
}

func (q *InMemoryQueue) ProcessNext(ctx context.Context) error {
	taskToProcess, err := q.Pop()
	if err != nil {
		if taskToProcess != nil {
			q.erroredTasks = append(q.erroredTasks, *taskToProcess)
		} else {
			return err
		}
	}
	queue_executor := executor.GetDefaultExecutor()
	res, err := queue_executor.Process(taskToProcess, ctx)
	slog.Info("Task completed",
		"task_id", taskToProcess.ID(),
		"result", res,
	)
	if err != nil {
		q.erroredTasks = append(q.erroredTasks, *taskToProcess)
		return err
	}
	q.completedTasks = append(q.completedTasks, *taskToProcess)
	return nil
}
