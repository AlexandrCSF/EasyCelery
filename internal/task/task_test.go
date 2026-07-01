package task

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTAsk(t *testing.T) {
	test_func := func(ctx context.Context) (any, error) {
		return 1, nil
	}

	actual := NewTask(test_func)
	expected := &Task{
		id:        actual.ID(),
		CreatedAt: actual.CreatedAt,
		fn:        test_func,
		status:    StatusPlan,
	}
	assert.Equal(t, expected, actual)
}
