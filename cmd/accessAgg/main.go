package main

import (
	"accessAggregator/internal/accesslog"
	"fmt"
	"time"
)

func main() {
	// TODO: tolerate common log rotation

	flags := parseFlags()
	ss := accesslog.Summaries{}
	handleShutdownSignal(&ss)

	c := make(chan accesslog.Record, len(flags.Files))
	ss := accesslog.Summaries{}

	for _, file := range flags.Files {
		go streamFileRecords(c, file, flags.fromStart)
	}
	go aggregateAndPrintSummaries(c, &ss, flags.Interval)

	for {
		ss.Print()
		fmt.Println()
		time.Sleep(flags.Interval)
	}
}
