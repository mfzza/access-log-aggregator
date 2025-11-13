package config

import (
	"flag"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantFiles     []string
		wantFromStart bool
		wantInterval  time.Duration
		wantError     string
	}{
		{
			name:          "single file",
			args:          []string{"-file", "app.log"},
			wantFiles:     []string{"app.log"},
			wantFromStart: false,
			wantInterval:  10 * time.Second,
		},
		{
			name:          "multiple files",
			args:          []string{"-file", "app.log", "-file", "error.log"},
			wantFiles:     []string{"app.log", "error.log"},
			wantFromStart: false,
			wantInterval:  10 * time.Second,
		},
		{
			name:          "with from-start",
			args:          []string{"-file", "app.log", "-from-start"},
			wantFiles:     []string{"app.log"},
			wantFromStart: true,
			wantInterval:  10 * time.Second,
		},
		{
			name:          "custom interval",
			args:          []string{"-file", "app.log", "-interval", "5s"},
			wantFiles:     []string{"app.log"},
			wantFromStart: false,
			wantInterval:  5 * time.Second,
		},
		{
			name:      "duplicate file",
			args:      []string{"-file", "app.log", "-file", "app.log"},
			wantError: "duplicate file: app.log",
		},
		{
			name:      "no files provided",
			args:      []string{},
			wantError: "missing required flag: at least one -file must be provided",
		},
		{
			name:      "invalid interval",
			args:      []string{"-file", "app.log", "-interval", "invalid"},
			wantError: "invalid value", // flag package error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag state for each test
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

			// suppress flag help output
			flag.CommandLine.SetOutput(io.Discard)

			os.Args = append([]string{"test"}, tt.args...)

			flags, err := ParseFlags()

			if tt.wantError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.wantError)
				} else if !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.wantError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(flags.Files, tt.wantFiles) {
				t.Errorf("Files = %v, want %v", flags.Files, tt.wantFiles)
			}
			if flags.FromStart != tt.wantFromStart {
				t.Errorf("fromStart = %v, want %v", flags.FromStart, tt.wantFromStart)
			}
			if flags.Interval != tt.wantInterval {
				t.Errorf("Interval = %v, want %v", flags.Interval, tt.wantInterval)
			}
		})
	}
}
