package main

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/logreader"
	"fmt"
	"io"
	"os"
	"time"
)

func streamFileRecords(c chan<- accesslog.Record, file string, fromStart bool) {

	r, err := logreader.NewReader(file, fromStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	for {
		rawRecord, err := r.GetRawRecord()
		if err == io.EOF {
			// break
			continue
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		record, err := accesslog.NewRecord(rawRecord)
		if err != nil {
			// ignore error/malformed/missing field
			// fmt.Fprintf(os.Stderr, "skipped line: %v\n", err)
			continue
		}
		// ss.AddRecord(record)
		c <- *record
	}
}

func aggregateAndPrintSummaries(c <-chan accesslog.Record, ss *accesslog.Summaries, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case r := <-c:
            ss.AddRecord(&r)
        case <-ticker.C:
            ss.Print()
        }
    }
}
