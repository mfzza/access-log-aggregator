package main

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/logreader"
	"fmt"
	"io"
	"os"
)

func processFiles(files []string) accesslog.Summaries {
	ss := accesslog.Summaries{}

	// TODO: replace this loop with goroutines
	for _, file := range files {
		r, err := logreader.NewReader(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
			os.Exit(1)
		}
		for {
			line, err := r.ReadLine()
			if err == io.EOF {
				break
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
			ss.AddRecord(record)
		}
		r.Close()
	}
	return ss
}
