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

func streamLogFile(fpath string, fromStart bool, ctx context.Context, rawRecords chan<- []byte) error {
	tf, err := tailer.NewTailFile(fpath, tailer.OsFS{}, fromStart, 200*time.Millisecond)
	if err != nil {
		return err
	}
	defer tf.Close()

	return runStreamLoop(tf, ctx, rawRecords)
}

func runStreamLoop(tf tailer.Tailer, ctx context.Context, rawRecords chan<- []byte) error {
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
				return fmt.Errorf("reading record: %w", err)
			}
			select {
			case rawRecords <- rawRecord:
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func aggregateAndPrintSummaries(ss accesslog.Summarieses, flags *Flags, rawRecords <-chan []byte, errCh <-chan error) bool {
	ticker := time.NewTicker(flags.Interval)
	defer ticker.Stop()

	var errs []error

	brokenRecord := 0
	printSummaries := func() {
		ss.Print()
		if brokenRecord > 0 {
			fmt.Println(" Missing field or Malformed log:", brokenRecord)
		}
	}

	printErrors := func(errs []error) {
		fmt.Println("\nFile error summary:")
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	// HACK: dirty fix for instant oneshot
	time.AfterFunc(10*time.Millisecond, printSummaries)

	for {
		select {
		// one of goroutine of tailed file error
		case err, ok := <-errCh:
			if ok {
				errs = append(errs, err)
				fmt.Fprintln(os.Stderr, err)

				// and total errors == total files
				if len(errs) == len(flags.Files) {
					printErrors(errs)
					return false
				}
			}

		// channel rawRecords receive from streamLogFile()
		case r, ok := <-rawRecords:
			// but rawRecords channel is closed
			if !ok {
				printSummaries()
				return true
			}

			// and rawRecords channel have data
			record, err := accesslog.NewRecord(r)
			if err != nil {
				brokenRecord++
				continue
			}
			ss.AddRecord(record)

		// periodically
		case <-ticker.C:
			printSummaries()
		}
	}
}
