package main

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/tailer"
	"context"
	"fmt"
	"io"
	"time"
)

func streamLogFile(fpath string, fromStart bool, ctx context.Context, rawRecords chan<- []byte) error {
	f, err := tailer.NewOSFile(fpath, fromStart)
	if err != nil {
		return fmt.Errorf("Failed to open file: %w", err)
	}
	tf, err := tailer.NewTailFile(fpath, f, 200*time.Millisecond)
	if err != nil {
		return err
	}
	defer tf.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			rawRecord, err := tf.GetRawRecord()
			if err == io.EOF {
				continue
			}
			if err != nil {
				return fmt.Errorf("failed to read record: %w", err)
			}
			select {
			case rawRecords <- rawRecord:
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func aggregateAndPrintSummaries(ss *accesslog.Summaries, interval time.Duration, rawRecords <-chan []byte) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case r, ok := <-rawRecords:
			if !ok {
				ss.Print()
				return
			}
			record, err := accesslog.NewRecord(r)
			if err != nil {
				continue
			}
			ss.AddRecord(record)

		case <-ticker.C:
			ss.Print()
		}
	}
}
