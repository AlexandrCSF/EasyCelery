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
	q := NewInMemoryQueue()

	assert.Equal(t, 0, q.Length())
	assert.False(t, q.HasNext())

	_, err := q.Pop()
	assert.Error(t, err)
}

func TestPushQueue(t *testing.T) {
	q := NewInMemoryQueue()

	q.Push(testTask())

	assert.Equal(t, 1, q.Length())
	assert.True(t, q.HasNext())
}

func TestPopQueue(t *testing.T) {
	q := NewInMemoryQueue()
	q.Push(testTask())

	popped, err := q.Pop()
	require.NoError(t, err)
	assert.NotNil(t, popped)
	assert.Equal(t, 0, q.Length())
	assert.False(t, q.HasNext())
}

func TestQueueFIFO(t *testing.T) {
	q := NewInMemoryQueue()

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
