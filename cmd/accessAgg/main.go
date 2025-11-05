package main

import (
	"accessAggregator/internal/accesslog"
	"fmt"
	"time"
)

func main() {
	// TODO: behaviour to default start read from tail (end of file), and read from beginning when have `-from-start` flag
	// TODO: tolerate common log rotation
	// TODO: tail -F like behaviour

	cfg := parseFlags()
	ss := accesslog.Summaries{}
	signalExit(ss)

	c := make(chan accesslog.Record, len(cfg.Files))

	for _, file := range cfg.Files {
		go processFiles(c, file)
	}
	go aggregateRecord(c, &ss)

	for {
		ss.Print()
		fmt.Println()
		time.Sleep(cfg.Interval)
	}
}
