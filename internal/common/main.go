package main

import (
	"easycelery/internal/queue"
	"easycelery/internal/runner"
	"flag"
	"log/slog"
	"os"
)

func main() {
	concurency := flag.Int("concurrency", 10, "Number of concurrent processes")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	mainQueue := queue.NewInMemoryQueue()
	mainRunner := runner.NewDefaultRunner(mainQueue, *concurency)
	mainRunner.RunExecutionForever()
}
