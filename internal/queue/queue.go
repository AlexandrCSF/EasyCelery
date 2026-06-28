package queue

import "easycelery/internal/task"

type Queue struct {
	tasksToProcess []task.Task
	completedTasks []task.Task
	erroredTasks   []task.Task
}

func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Push(task task.Task) {
	q.tasksToProcess = append(q.tasksToProcess, task)
}

func (q *Queue) ProcessTask() {
	taskToProcess := q.tasksToProcess[0]
	res, err := taskToProcess.Process()
}
