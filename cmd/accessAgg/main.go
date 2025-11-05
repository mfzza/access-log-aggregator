package main

import (
	"accessAggregator/internal/accesslog"
	"fmt"
	"time"
)

func main() {
	// TODO: tolerate common log rotation

	cfg := parseFlags()
	ss := accesslog.Summaries{}
	signalExit(ss)

	c := make(chan accesslog.Record, len(cfg.Files))

	for _, file := range cfg.Files {
		go processFiles(c, file, cfg.fromStart)
	}
	go aggregateRecord(c, &ss)

	for {
		ss.Print()
		fmt.Println()
		time.Sleep(cfg.Interval)
	}
}
