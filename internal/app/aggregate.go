package app

import (
	"accessAggregator/internal/accesslog"
	"accessAggregator/internal/config"
	"fmt"
	"io"
	"time"
)

func aggr(aggrDone chan<- struct{}, flags config.Flags, data <-chan []byte, summaries accesslog.Summarizer, out io.Writer) {
	ticker := time.NewTicker(flags.Interval)
	defer ticker.Stop()

	var malformRecord int
	printSummaries := func() {
		fmt.Fprint(out, summaries.Format())
		if malformRecord > 0 {
			fmt.Fprintln(out, yellow+"missing field or malformed log:", malformRecord, reset)
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
				fmt.Fprint(out, green+"\nPrinting final summary:"+reset)
				printSummaries()
				close(aggrDone)
				return
			}
			if err := summaries.Aggregate(r); err != nil {
				malformRecord++
				continue
			}
		}
	}
}
