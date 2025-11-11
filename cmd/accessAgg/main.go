package main

import (
	"accessAggregator/internal/accesslog"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	flags, err := parseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		flag.Usage()
		os.Exit(2)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	rawRecord := make(chan []byte, 100)
	ss := accesslog.Summaries{}

	var tailWG sync.WaitGroup

	for _, file := range flags.Files {
		tailWG.Go(func() {
			if err := streamLogFile(file, flags.fromStart, ctx, rawRecord); err != nil {
				fmt.Printf("Error tailing %s: %v\n", file, err)
			}
		})
	}

	var aggWG sync.WaitGroup

	aggWG.Go(func() {
		aggregateAndPrintSummaries(&ss, flags.Interval, rawRecord)
	})

	sig := <-done
	cancel()
	tailWG.Wait()
	close(errCh)
	close(rawRecord)

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Printf("Received signal: %s.\n", sig)
	fmt.Println("Gracefully shutting down... Printing final summary")

	aggWG.Wait()

	wg.Wait()
	os.Exit(0)
}
