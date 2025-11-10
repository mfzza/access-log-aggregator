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

func parseFlags() (Flags, error) {
	var flags Flags
	seen := map[string]bool{}

	flag.Func("file", "path to log file", func(file string) error {
		if seen[file] {
			return fmt.Errorf("duplicate file: %s", file)
		}
		seen[file] = true
		flags.Files = append(flags.Files, file)
		return nil
	})

	flag.BoolVar(&flags.fromStart, "from-start", false, "read from beginning")
	flag.DurationVar(&flags.Interval, "interval", 10*time.Second, "summary interval")

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return Flags{}, err
	}

	if len(flags.Files) == 0 {
		return Flags{}, fmt.Errorf("missing required flag: at least one -file must be provided")
	}

	return flags, nil
}
