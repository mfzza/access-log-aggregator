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

	var wg sync.WaitGroup

	for _, file := range flags.Files {
		wg.Go(func() {
			if err := streamLogFile(file, flags.fromStart, ctx, rawRecord); err != nil {
				fmt.Printf("Error tailing %s: %v\n", file, err)
			}
		})
	}
	wg.Go(func() {
		aggregateAndPrintSumaries(&ss, flags.Interval, ctx, rawRecord)
	})

	sig := <-done
	cancel()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Printf("Received signal: %s.\n", sig)
	fmt.Println("Gracefully shutting down... Printing final summary")
	ss.Print()
	wg.Wait()
	os.Exit(0)
	fmt.Println("EXIT")
}
