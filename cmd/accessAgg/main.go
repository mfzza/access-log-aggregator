package main

import (
	"accessAggregator/internal/app"
	"accessAggregator/internal/config"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flags, err := config.ParseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		flag.Usage()
		os.Exit(2)
	}

	// ctx, cancel := context.WithCancel(context.Background())
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, flags); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

}
