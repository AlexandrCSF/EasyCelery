package runner

import (
	"context"
	"easycelery/internal/queue"
	"easycelery/internal/task"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func BenchmarkRunner(b *testing.B) {

}
func TestRunnerSingleTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := queue.NewInMemoryQueue()
	r := NewDefaultRunnerDefaultValues(q)

	var executed atomic.Int32
	r.SendTask(*task.NewTask(func(ctx context.Context) (any, error) {
		executed.Add(1)
		return "ok", nil
	}))

	go r.Run(ctx)

	require.Eventually(t, func() bool {
		return executed.Load() == 1
	}, time.Second, 10*time.Millisecond)
}

func TestRunnerMultipleTasks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := queue.NewInMemoryQueue()
	r := NewDefaultRunnerDefaultValues(q)

	const taskCount = 3
	var executed atomic.Int32

	for range taskCount {
		r.SendTask(*task.NewTask(func(ctx context.Context) (any, error) {
			executed.Add(1)
			return nil, nil
		}))
	}

	go r.Run(ctx)

	require.Eventually(t, func() bool {
		return executed.Load() == taskCount
	}, 2*time.Second, 10*time.Millisecond)
}

func TestRunnerContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	q := queue.NewInMemoryQueue()
	r := NewDefaultRunnerDefaultValues(q)

	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runner did not stop after context cancellation")
	}
}
