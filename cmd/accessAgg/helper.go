package main

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/tailer"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

func streamLogFile(fpath string, fromStart bool, ctx context.Context, rawRecords chan<- []byte) {
	f, err := tailer.NewOSFile(fpath, fromStart)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tf, err := tailer.NewTailFile(fpath, f, 200*time.Millisecond)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer tf.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			rawRecord, err := tf.GetRawRecord()
			if err == io.EOF {
				continue
			}
			if err != nil {
				return
			}
			select {
			case rawRecords <- rawRecord:
			case <-ctx.Done():
				return
			}
		}
	}
}

func aggregateAndPrintSumaries(ss *accesslog.Summaries, interval time.Duration, ctx context.Context, rawRecords <-chan []byte) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case r := <-rawRecords:
			record, err := accesslog.NewRecord(r)
			if err != nil {
				continue
			}
			ss.AddRecord(record)
		case <-ticker.C:
			ss.Print()
		case <-ctx.Done():
			ss.Print()
			return
		}
	}
}
