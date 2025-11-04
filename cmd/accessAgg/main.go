package main

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/logreader"
	"flag"
	"fmt"
	"io"
	"os"
)

// func NewAccessLog(time time.Time, host string, statusCode int, duration float64) *accessLog {
// 	return &accessLog{time, host, statusCode, duration}
// }

func main() {
	// TODO: implement multiple file read with goroutine
	// TODO: behaviour to default start read from tail (end of file), and read from beginning when have `-from-start` flag
	// TODO: tolerate common log rotation
	// TODO: -interval flag
	// TODO: tail -F like behaviour
	// TODO: handle graceful exit

	var files []string
	flag.Func("file", "path to log file (required)\ncan be specified multiple time, example: -file a.log -file b.log",
		func(file string) error {
			files = append(files, file)
			return nil
		})
	flag.Parse()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "missing required flag: at least one -file <path> must be provided")
		flag.Usage()
		os.Exit(2)
	}

	ss := accesslog.Summaries{}

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
	ss.Print()
}
