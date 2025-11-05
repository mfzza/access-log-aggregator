package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Flags struct {
	Files     []string
	fromStart bool
	Interval  time.Duration
}

func parseFlags() Flags {
	var flags Flags
	seen := map[string]bool{}
	flag.Func("file", "path to log file (required)\ncan be specified multiple time, example: -file a.log -file b.log",
		func(file string) error {
			if seen[file] {
				return fmt.Errorf("duplicate file: %s", file) // or emit error?
				// return nil // skip
			}
			seen[file] = true
			flags.Files = append(flags.Files, file)
			return nil

		})

	flag.BoolVar(&flags.fromStart, "from-start", false, "optional: read from beginning (default: tail from end)")
	flag.DurationVar(&flags.Interval, "interval", 10*time.Second, "optional: summary interval")

	flag.Parse()
	if len(flags.Files) == 0 {
		fmt.Fprintln(os.Stderr, "missing required flag: at least one -file <path> must be provided")
		flag.Usage()
		os.Exit(2)
	}
	return flags
}

