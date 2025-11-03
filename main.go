package main

import (
	"accessAggregator/internal/accesslog"
	"bufio"
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

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening a file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	r := bufio.NewReader(file)

	ss := accesslog.Summaries{}
	for i := 1; ; i++ {
		// use scanner?
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "cant read file: %v\n", err)
			break
		}
		record, err := accesslog.NewRecord(line)
		if err != nil {
			fmt.Println(err)
			continue
		}
		ss.AddRecord(record)
	}

	ss.Print()
}
