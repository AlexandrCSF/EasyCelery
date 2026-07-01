package task

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
