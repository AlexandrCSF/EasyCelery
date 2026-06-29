package main

import (
	"easycelery/internal/queue"
	"easycelery/internal/runner"
	"easycelery/internal/task"
	"flag"
	"log/slog"
	"math/rand"
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
	mainRunner.SendTask(*task.NewTask(addRandom, nil))

	mainRunner.RunExecutionForever()

}

func addRandom() (any, error) {
	return rand.Int() + rand.Int(), nil
}
