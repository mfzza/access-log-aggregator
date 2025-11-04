package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Conf struct {
	Files     []string
	fromStart bool
	Interval  time.Duration
}

func parseFlags() Conf {
	var cfg Conf
	seen := map[string]bool{}
	flag.Func("file", "path to log file (required)\ncan be specified multiple time, example: -file a.log -file b.log",
		func(file string) error {
			if seen[file] {
				return fmt.Errorf("duplicate file: %s", file) // or emit error?
				// return nil // skip
			}
			seen[file] = true
			cfg.Files = append(cfg.Files, file)
			return nil

		})

	flag.BoolVar(&cfg.fromStart, "from-start", false, "optional: read from beginning (default: tail from end)")
	flag.DurationVar(&cfg.Interval, "interval", 10*time.Second, "optional: summary interval")

	flag.Parse()
	if len(cfg.Files) == 0 {
		fmt.Fprintln(os.Stderr, "missing required flag: at least one -file <path> must be provided")
		flag.Usage()
		os.Exit(2)
	}
	return cfg
}

