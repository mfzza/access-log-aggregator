package main

import (
	"accessAggregator/internal/accesslog"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	flags := parseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	rawRecord := make(chan []byte, len(flags.Files))
	ss := accesslog.Summaries{}

	var wg sync.WaitGroup

	for _, file := range flags.Files {
		wg.Go(func () {
			streamLogFile(file, flags.fromStart, ctx, rawRecord)
		})
	}
	// FIXME: first run should print instantly
	wg.Go(func ()  {
		aggregateAndPrintSumaries(&ss, flags.Interval, ctx, rawRecord)
	})

	sig := <-done
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Printf("Received signal: %s.\n", sig)
	fmt.Println("Gracefully shutting down... Printing final summary")

	cancel()
	wg.Wait()
}
