package main

import (
	"accessAggregator/internal/accesslog"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func signalExit(ss accesslog.Summaries){
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Printf("Received signal: %s.\n", sig)
		fmt.Println(" Gracefully shutting down...Printing final summary")
		ss.Print()
		os.Exit(0)
	}()

}
