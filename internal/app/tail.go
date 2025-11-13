package app

import (
	"accessAggregator/internal/tailer"
	"context"
	"fmt"
	"io"
	"time"
)

func tail(ctx context.Context, fpath string, fromStart bool, data chan<- []byte) error {
	tf, err := tailer.NewTailFile(fpath, tailer.OsFS{}, fromStart)
	if err != nil {
		return err
	}
	return streamLoop(tf, ctx, data)
}

func streamLoop(tf tailer.Tailer, ctx context.Context, data chan<- []byte) error {
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
