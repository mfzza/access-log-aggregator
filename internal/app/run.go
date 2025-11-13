package app

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/config"
	"context"
	"fmt"
	"os"
	"sync"
)

func Run(ctx context.Context, flags config.Flags) error {
    summaries := accesslog.NewSummaries()
    data := make(chan []byte, 100)

	// producer
    var wg sync.WaitGroup
    for _, file := range flags.Files {
        wg.Go(func() {
            if err := tail(ctx, file, flags.FromStart, data); err != nil {
                fmt.Fprintf(os.Stderr, "[%s] error: %v\n", file, err)
            }
        })
    }

	// consumer
    aggrDone := make(chan struct{})
    go aggr(aggrDone, flags, data, summaries, os.Stdout)

    wg.Wait()

    close(data)
    <-aggrDone

    fmt.Println("Gracefully shut down...")
    return nil
}
