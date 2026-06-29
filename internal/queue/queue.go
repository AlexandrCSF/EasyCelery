package queue

import (
	"easycelery/internal/task"
	"errors"
	"log/slog"
)

type Queue interface {
	Push(task task.Task)
	Pop() (*task.Task, error)
	ProcessNext() error
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

func (q *InMemoryQueue) Pop() (*task.Task, error) {
	if len(q.tasksToProcess) == 0 {
		return nil, errors.New("queue is empty")
	}
	poppedTask := q.tasksToProcess[0]
	q.tasksToProcess = q.tasksToProcess[1:]
	return &poppedTask, nil
}

func (q *InMemoryQueue) ProcessNext() error {
	taskToProcess, err := q.Pop()
	if err != nil {
		if taskToProcess != nil {
			q.erroredTasks = append(q.erroredTasks, *taskToProcess)
		} else {
			return err
		}
	}
	res, err := taskToProcess.Process()
	slog.Info("Task", taskToProcess.Id(), "completed with result: ", res)
	if err != nil {
		q.erroredTasks = append(q.erroredTasks, *taskToProcess)
		return err
	}
	q.completedTasks = append(q.completedTasks, *taskToProcess)
	return nil
}
