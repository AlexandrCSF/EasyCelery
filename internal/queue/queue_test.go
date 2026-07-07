package queue

import (
	"context"
	"easycelery/internal/task"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkQueue(b *testing.B) {

}
func testTask() *task.Task {
	return task.NewTask(func(ctx context.Context) (any, error) {
		return 1, nil
	})
}

func TestEmptyQueue(t *testing.T) {
	q := NewInMemoryQueue(1)

	assert.Equal(t, 0, q.Length())

	_, err := q.Pop()
	assert.Error(t, err)
}

func TestPushQueue(t *testing.T) {
	q := NewInMemoryQueue(1)

	q.Push(testTask())

	assert.Equal(t, 1, q.Length())
}

func TestPopQueue(t *testing.T) {
	q := NewInMemoryQueue(1)
	q.Push(testTask())

	popped, err := q.Pop()
	require.NoError(t, err)
	assert.NotNil(t, popped)
	assert.Equal(t, 0, q.Length())
}

func TestQueueFIFO(t *testing.T) {
	q := NewInMemoryQueue(1)

	first := testTask()
	second := testTask()
	third := testTask()

	q.Push(first)
	q.Push(second)
	q.Push(third)

	poppedFirst, err := q.Pop()
	require.NoError(t, err)
	assert.Equal(t, first.ID(), poppedFirst.ID())

	poppedSecond, err := q.Pop()
	require.NoError(t, err)
	assert.Equal(t, second.ID(), poppedSecond.ID())

	poppedThird, err := q.Pop()
	require.NoError(t, err)
	assert.Equal(t, third.ID(), poppedThird.ID())

	assert.Equal(t, 0, q.Length())
}
