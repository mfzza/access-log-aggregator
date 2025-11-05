package main

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/logreader"
	"fmt"
	"io"
	"os"
)

func processFiles(c chan<- accesslog.Record, file string) {

	r, err := logreader.NewReader(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			// break
			continue
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		record, err := accesslog.NewRecord(line)
		if err != nil {
			// ignore error/malformed/missing field
			// fmt.Fprintf(os.Stderr, "skipped line: %v\n", err)
			continue
		}
		// ss.AddRecord(record)
		c <- *record
	}
}

func aggregateRecord(c <-chan accesslog.Record, ss *accesslog.Summaries) {
	for r := range c {
		ss.AddRecord(&r)
	}
}
