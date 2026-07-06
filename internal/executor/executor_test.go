package executor

import (
	"context"
	"easycelery/internal/task"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkExecutor(b *testing.B) {

}
func TestExecutorSuccessfulExecution(t *testing.T) {
	ex := GetDefaultExecutor()
	tsk := task.NewTask(func(ctx context.Context) (any, error) {
		return "result", nil
	})

	result, err := ex.Process(tsk, context.Background())
	require.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, "result", tsk.GetResult())
	assert.Equal(t, task.StatusCompleted, tsk.Status())
	assert.False(t, tsk.CompletedAt().IsZero())
}

func TestExecutorErrorHandling(t *testing.T) {
	ex := GetDefaultExecutor()
	expectedErr := errors.New("execution failed")
	tsk := task.NewTask(func(ctx context.Context) (any, error) {
		return nil, expectedErr
	})

	result, err := ex.Process(tsk, context.Background())
	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
	assert.Equal(t, task.StatusError, tsk.Status())
	assert.Equal(t, expectedErr.Error(), tsk.Error)
	assert.False(t, tsk.CompletedAt().IsZero())
}

func TestExecutorStatusTransitions(t *testing.T) {
	ex := GetDefaultExecutor()

	t.Run("plan to processing to completed", func(t *testing.T) {
		tsk := task.NewTask(func(ctx context.Context) (any, error) {
			return "done", nil
		})
		assert.Equal(t, task.StatusPlan, tsk.Status())

		_, err := ex.Process(tsk, context.Background())
		require.NoError(t, err)
		assert.Equal(t, task.StatusCompleted, tsk.Status())
		assert.False(t, tsk.StartAt().IsZero())
	})

	t.Run("plan to processing to error", func(t *testing.T) {
		tsk := task.NewTask(func(ctx context.Context) (any, error) {
			return nil, errors.New("fail")
		})
		assert.Equal(t, task.StatusPlan, tsk.Status())

		_, err := ex.Process(tsk, context.Background())
		require.Error(t, err)
		assert.Equal(t, task.StatusError, tsk.Status())
		assert.False(t, tsk.StartAt().IsZero())
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		tsk := task.NewTask(func(ctx context.Context) (any, error) {
			return "should not run", nil
		})

		_, err := ex.Process(tsk, ctx)
		require.Error(t, err)
		assert.Equal(t, task.StatusError, tsk.Status())
		assert.Equal(t, context.Canceled.Error(), tsk.Error)
	})
}

func TestExecutorSetsTimestamps(t *testing.T) {
	ex := GetDefaultExecutor()
	before := time.Now()

	tsk := task.NewTask(func(ctx context.Context) (any, error) {
		return nil, nil
	})

	_, err := ex.Process(tsk, context.Background())
	require.NoError(t, err)

	after := time.Now()
	assert.False(t, tsk.StartAt().Before(before))
	assert.False(t, tsk.StartAt().After(after))
	assert.False(t, tsk.CompletedAt().Before(before))
	assert.False(t, tsk.CompletedAt().After(after))
}
