package main

import (
	"accessAggregator/internal/accesslog"
	"sync"
)

func main() {
	// TODO: behaviour to default start read from tail (end of file), and read from beginning when have `-from-start` flag
	// TODO: tolerate common log rotation
	// TODO: -interval flag
	// TODO: tail -F like behaviour

	cfg := parseFlags()
	ss := accesslog.Summaries{}
	signalExit(ss)

	c := make(chan accesslog.Record, len(cfg.Files))

	done := make(chan struct{})
	go aggregateRecord(c, &ss, done)

	var wg sync.WaitGroup
	for _, file := range cfg.Files {
		wg.Add(1)
		go processFiles(c, file, &wg)
	}

	wg.Wait()
	close(c)

	<-done
	ss.Print()
}
