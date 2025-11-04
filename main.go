package main

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/logreader"
	"fmt"
	"io"
	"os"
)

// func NewAccessLog(time time.Time, host string, statusCode int, duration float64) *accessLog {
// 	return &accessLog{time, host, statusCode, duration}
// }

func main() {
	// TODO: implement multiple file read with goroutine
	// TODO: implement flag parameter
	// TODO: behaviour to default start read from tail (end of file), and read from beginning when have `-from-start` flag
	// TODO: tolerate common log rotation
	// TODO: -interval flag
	// TODO: tail -F like behaviour
	// TODO: handle graceful exit

	r, err := logreader.NewReader(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	ss := accesslog.Summaries{}
	for i := 1; ; i++ {

		line, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}

		record, err := accesslog.NewRecord(line)
		if err != nil {
			fmt.Println("Invalid json format, skipped line")
			continue
		}
		ss.AddRecord(record)
	}
	ss.Print()
}
