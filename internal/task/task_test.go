package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	testFunc := func(ctx context.Context) (any, error) {
		return 1, nil
	}

	before := time.Now()
	actual := NewTask(testFunc)
	after := time.Now()

	assert.NotEmpty(t, actual.ID())
	assert.Equal(t, StatusPlan, actual.Status())
	assert.False(t, actual.CreatedAt.Before(before))
	assert.False(t, actual.CreatedAt.After(after))

	result, err := actual.Func()(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestTaskSuccessfulExecution(t *testing.T) {
	tsk := NewTask(func(ctx context.Context) (any, error) {
		return "ok", nil
	})

	result, err := tsk.Func()(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "ok", result)
}

func TestTaskExecutionWithError(t *testing.T) {
	expectedErr := errors.New("task failed")
	tsk := NewTask(func(ctx context.Context) (any, error) {
		return nil, expectedErr
	})

	result, err := tsk.Func()(context.Background())
	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestTaskStatusUpdate(t *testing.T) {
	tsk := NewTask(func(ctx context.Context) (any, error) {
		return nil, nil
	})

	assert.Equal(t, StatusPlan, tsk.Status())

	tsk.SetStatus(StatusProcessing)
	assert.Equal(t, StatusProcessing, tsk.Status())

	tsk.SetStatus(StatusCompleted)
	assert.Equal(t, StatusCompleted, tsk.Status())

	tsk.SetStatus(StatusError)
	assert.Equal(t, StatusError, tsk.Status())
}

func TestTaskResultSaving(t *testing.T) {
	tsk := NewTask(func(ctx context.Context) (any, error) {
		return 42, nil
	})

	result, err := tsk.Func()(context.Background())
	require.NoError(t, err)

	tsk.Result = result
	tsk.SetStatus(StatusCompleted)

	assert.Equal(t, 42, tsk.Result)
	assert.Equal(t, StatusCompleted, tsk.Status())
}
