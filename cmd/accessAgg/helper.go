package main

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/fileutil"
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

func streamFileRecords(c chan<- accesslog.Record, fpath string, fromStart bool) {

	tailFile, err := fileutil.NewTailFile(fpath, fromStart)
	if err != nil {
		fmt.Println(err)
		// exit since the log file could not be opened or accessed.
		os.Exit(1)
	}

	for {
		rawRecord, err := tailFile.NextLine()
		if err == io.EOF {
			// break
			continue
		}
		if err != nil {
			// TODO: what should we do for this error?
			// this error is happen when it failed to readBytes("\n"),
			// or detect log rotation failed
			fmt.Fprintf(os.Stderr, "%v\n", err)
			// os.Exit(1)
			continue
		}

		record, err := accesslog.NewRecord(rawRecord)
		if err != nil {
			// ignore error/malformed/missing field
			// TODO: log?
			// fmt.Fprintf(os.Stderr, "skipped line: %v\n", err)
			continue
		}
		c <- *record
	}
}

func streamFileRecords2(c chan<- accesslog.Record, fpath string, fromStart bool) {
	tail, err := os.Open(fpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file: %v\n", err)
		os.Exit(1)
	}

	rotate, err := fileutil.NewLogRotate(fpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, " %v\n", err)
		os.Exit(1)
	}

	if !fromStart {
		if _, err := tail.Seek(0, io.SeekEnd); err != nil {
			fmt.Fprintf(os.Stderr, "failed to seek: %v\n", err)
			os.Exit(1)
		}
	}

	r := bufio.NewReader(tail)

	for {
		line, err := fileutil.NextLine(r)
		if err == io.EOF {
			time.Sleep(200 * time.Millisecond)
			// handle EOF
			rotate.CheckRotation(fpath)
			if rotate.Truncated {
				if _, err := tail.Seek(0, io.SeekStart); err != nil {
					fmt.Fprintf(os.Stderr, "failed to seek: %v", err)
					os.Exit(1)
				}
				rotate.Truncated = false
			}
			if rotate.Renamed {
				tail.Close()
				tail, err = os.Open(fpath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "cant open new log file: %v", err)
					os.Exit(1)
				}
				r.Reset(tail)
				rotate.Renamed = false
			}
			continue
		}
		if err != nil {
			// TODO: what should we do for this error?
			// this error is happen when it failed to readBytes("\n"),
			// or detect log rotation failed
			// fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		record, err := accesslog.NewRecord(line)
		if err != nil {
			continue
		}
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
