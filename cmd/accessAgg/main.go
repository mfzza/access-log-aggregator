package main

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/config"
	"accessAggregator/internal/tailer"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	flags, err := config.ParseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		flag.Usage()
		os.Exit(2)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		fmt.Println("\n\n\nreceive signal:", sig)
		cancel()
	}()

	summaries := accesslog.NewSummaries()
	data := make(chan []byte, 100)

	// producer
	var wg sync.WaitGroup
	for _, file := range flags.Files {
		// NOTE: https://appliedgo.net/spotlight/go-1.25-waitgroup-go/
		wg.Go(func() {
			if err := tail(ctx, file, flags.FromStart, data); err != nil {
				fmt.Fprintf(os.Stderr, "[%s] error tailing: %v", file, err)
			}
		})
	}

	// consumer
	final := make(chan struct{})
	go aggr(final, flags, data, summaries)

	// wait all tail finish, then close data channel, then wait for aggr end
	wg.Wait()
	close(data)
	<-final
	fmt.Println("Gracefully shutting down...")
}

func tail(ctx context.Context, fpath string, fromStart bool, data chan<- []byte) error {
	tf, err := tailer.NewTailFile(fpath, tailer.OsFS{}, fromStart)
	if err != nil {
		return err
	}
	defer tf.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			rawRecord, err := tf.GetRawRecord()
			if err == io.EOF {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}
			if err != nil {
				return fmt.Errorf("reading record: %w", err)
			}
			data <- rawRecord
		}
	}

}

func aggr(aggrEnd chan<- struct{}, flags config.Flags, data <-chan []byte, summaries accesslog.Summarizer) {
	ticker := time.NewTicker(flags.Interval)
	defer ticker.Stop()

	var malformRecord int
	printSummaries := func() {
		fmt.Fprint(os.Stdout, summaries.Format())
		if malformRecord > 0 {
			fmt.Fprintln(os.Stdout, "missing field or malformed log: ", malformRecord)
		}
	}

	for {
		select {
		case <-ticker.C:
			printSummaries()

		// keep process data even after context canceled
		// to drain remaining data, then give signal
		// when channel already empty
		case r, ok := <-data:
			if !ok {
				fmt.Println("Printing final summary")
				printSummaries()
				close(aggrEnd)
				return
			}
			record, err := accesslog.NewRecord(r)
			if err != nil {
				malformRecord++
				continue
			}
			summaries.AddRecord(record)
		}
	}
}
