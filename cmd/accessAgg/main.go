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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, len(flags.Files))
	rawRecords := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())

	ss := accesslog.NewSummaries()

	var tailWG sync.WaitGroup

	for _, file := range flags.Files {
		tailWG.Go(func() {
			if err := streamLogFile(file, flags.fromStart, ctx, rawRecords); err != nil {
				errorStream := fmt.Errorf("[%s] error tailing: %w", file, err)
				errCh <- errorStream
			}
		})
	}

	exitCh := make(chan struct{})
	var aggWG sync.WaitGroup

	aggWG.Go(func() {
		if ok := aggregateAndPrint(ss, flags, rawRecords, errCh, os.Stdout, os.Stderr); !ok {
			exitCh <- struct{}{}
		}
	})

	select {
	case <-exitCh:
		fmt.Fprintln(os.Stderr, "failed to process all file, shutting down...")
		os.Exit(1)

	case sig := <-sigCh:
		fmt.Printf("\n\n\nReceived signal: %s.\n", sig)

		cancel()
		tailWG.Wait()

		fmt.Println("Gracefully shutting down... Printing final summary")

		close(rawRecords)

		aggWG.Wait()
	}
}
