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
	warning := "AHAHAHAHAHAHAHAHA"
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, warning)
		os.Exit(2)
	}

	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "Too many arguments\n"+warning)
		os.Exit(2)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening a file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	r := bufio.NewReader(file)

	als := make([]accesslog.Record, 0)
	for i := 1; ; i++ {
		// use scanner?
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("cant read, need to learn:", err)
			break
		}
		al, err := accesslog.NewRecord(line)
		if err != nil {
			fmt.Println(err)
			continue
		}

		als = append(als, *al)
		fmt.Print("=== ", i, " --- ", len(line))
		al.Print()
	}

	fmt.Println()
	fmt.Println()
	fmt.Println()

	ss := accesslog.Summaries{}

	s, _ := accesslog.NewSummary("chatgpt.com")
	ss = append(ss, *s)
	ss = append(ss, *s)
	ss.Print()





}
