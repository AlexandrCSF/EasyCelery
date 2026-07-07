package main

import (
	"context"
	"easycelery/internal/queue"
	"easycelery/internal/runner"
	"easycelery/internal/task"
	"flag"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	concurrency := flag.Int("concurrency", 5, "Number of concurrent processes")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	mainQueue := queue.NewInMemoryQueue(*concurrency)
	config := runner.Config{
		Workers: *concurrency,
	}
	mainRunner := runner.NewDefaultRunner(mainQueue, config)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mainRunner.Run(ctx)
	}()
	for i := 0; i < 10; i++ {
		mainQueue.Push(task.NewTask(func(ctx context.Context) (any, error) {
			time.Sleep(1 * time.Second)
			return sumTwo(10, 15)
		}))

	}

	<-ctx.Done()
	wg.Wait()
}

func addRandom() (any, error) {
	return rand.Intn(10) + rand.Intn(10), nil
}

func sumTwo(a, b int) (any, error) {
	return a + b, nil
}
