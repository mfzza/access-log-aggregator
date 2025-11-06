package main

import (
	"accessAggregator/internal/accesslog"
	"fmt"
	"io"
	"os"
	"time"
)

func streamFileRecords(c chan<- accesslog.Record, file string, fromStart bool) {

	f, err := os.Open(file)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		os.Exit(1)
	}
	defer f.Close()

	if !fromStart {
		SeekToLastNLines(f, 10)
	}

	r := accesslog.NewReader(f)

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

func SeekToLastNLines(f *os.File, n int) {
	if n <= 0 {
		f.Seek(0, io.SeekEnd)
		return
	}

	csize := 512
	stat, err := f.Stat()
	if err != nil {
		return
	}
	fsize := stat.Size()

	var linesFound int
	var offset int64

	for {
		if fsize-int64(csize)-offset < 0 {
			csize = int(fsize - offset)
		}

		offset += int64(csize)
		if offset > fsize {
			offset = fsize
		}

		// Move backward
		pos := max(fsize-offset, 0)
		f.Seek(pos, io.SeekStart)

		tmp := make([]byte, csize)
		f.Read(tmp)

		for i := len(tmp) - 1; i >= 0; i-- {
			if tmp[i] == '\n' {
				linesFound++
				if linesFound > n {
					f.Seek(pos+int64(i+1), io.SeekStart)
					return
				}
			}
		}

		if pos == 0 {
			f.Seek(0, io.SeekEnd)
			return
		}
	}
}

