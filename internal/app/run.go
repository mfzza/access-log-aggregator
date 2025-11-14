package app

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/config"
	"context"
	"fmt"
	"io"
	"sync"
)

func Run(ctx context.Context, flags config.Flags, out io.Writer, outErr io.Writer) error {
	summaries := accesslog.NewSummaries()

	// scale with * 25, but min 100 and max 10000
	bufSize := min(max(len(flags.Files)*25, 100), 10000)
	data := make(chan []byte, bufSize)

	// producer
	var wg sync.WaitGroup
	for _, file := range flags.Files {
		wg.Go(func() {
			if err := tail(ctx, file, flags.FromStart, data); err != nil {
				fmt.Fprintf(outErr, red+"[%s] error: %v\n"+reset, file, err)
			}
		})
	}

	// consumer
	aggrDone := make(chan struct{})
	go aggr(aggrDone, flags, data, summaries, out)

	wg.Wait()

	close(data)
	<-aggrDone

	fmt.Fprintln(out, "Gracefully shut down...")
	return nil
}
