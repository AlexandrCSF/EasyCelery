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
	HandleNext(ctx context.Context) error
	HasNext() bool
	GetExecutor() executor.Executor
	HandleTaskError(task task.Task) error
	HandleTaskSuccess(task task.Task) error
}

type InMemoryQueue struct {
	tasksToProcess []task.Task
	completedTasks []task.Task
	erroredTasks   []task.Task
	executor       executor.Executor
}

func (q *InMemoryQueue) HandleTaskError(task task.Task) error {
	q.erroredTasks = append(q.erroredTasks, task)
	return nil
}

func (q *InMemoryQueue) HandleTaskSuccess(task task.Task) error {
	q.completedTasks = append(q.completedTasks, task)
	return nil
}

func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		executor: executor.GetDefaultExecutor(),
	}
}
func (q *InMemoryQueue) GetExecutor() executor.Executor {
	return executor.GetDefaultExecutor()
}

func (q *InMemoryQueue) Push(task task.Task) {
	q.tasksToProcess = append(q.tasksToProcess, task)
}

func (q *InMemoryQueue) Length() int {
	return len(q.tasksToProcess)
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

func (q *InMemoryQueue) HandleNext(ctx context.Context) error {
	taskToProcess, err := q.Pop()
	if err != nil {
		if taskToProcess != nil {
			_ = q.HandleTaskError(*taskToProcess)
			return err
		} else {
			return err
		}
	}
	res, err := q.GetExecutor().Process(taskToProcess, ctx)
	slog.Info("Task completed",
		"task_id", taskToProcess.ID(),
		"result", res,
	)
	if err != nil {
		_ = q.HandleTaskError(*taskToProcess)
		return err
	}
	err = q.HandleTaskSuccess(*taskToProcess)
	if err != nil {
		return err
	}
	return nil
}
