package main

import (
	"accessAggregator/internal/accesslog"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// TODO: tolerate common log rotation

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	flags := parseFlags()

	c := make(chan accesslog.Record, len(flags.Files))
	ss := accesslog.Summaries{}

	for _, file := range flags.Files {
		go streamFileRecords(c, file, flags.fromStart)
	}
	// FIXME: first run should print instantly
	go aggregateAndPrintSummaries(c, &ss, flags.Interval)

	sig := <-done
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Printf("Received signal: %s.\n", sig)
	fmt.Println("Gracefully shutting down... Printing final summary")
	ss.Print()
}
