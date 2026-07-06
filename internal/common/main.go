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
	concurency := flag.Int("concurrency", 10, "Number of concurrent processes")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	mainQueue := queue.NewInMemoryQueue()
	mainRunner := runner.NewDefaultRunner(mainQueue, *concurency)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mainRunner.Run(ctx)
	}()
	for i := 0; i < 10; i++ {
		mainRunner.SendTask(task.NewTask(func(ctx context.Context) (any, error) {
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
