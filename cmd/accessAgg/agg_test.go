package main

import (
	"accessAggregator/internal/accesslog"
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

// mockSummarizer implements accesslog.Summarizer for testing
type mockSummarizer struct {
	records    []*accesslog.Record
	formatCall int
	formatStr  string
}

func (m *mockSummarizer) AddRecord(r *accesslog.Record) {
	m.records = append(m.records, r)
}

func (m *mockSummarizer) Format() string {
	m.formatCall++
	if m.formatStr != "" {
		return m.formatStr
	}
	return "Mock Summary Output"
}

// TestAggregateAndPrintSummaries tests the main aggregation function
func TestAggregateAndPrintSummaries(t *testing.T) {
	tests := []struct {
		name            string
		setupSummarizer func() *mockSummarizer
		flags           Flags
		rawRecords      [][]byte
		errChInput      []error
		wantOutput      []string
		wantErrOutput   []string
		wantReturn      bool
		setupTimeout    time.Duration
	}{
		{
			name: "normal operation with records and periodic output",
			setupSummarizer: func() *mockSummarizer {
				return &mockSummarizer{formatStr: "Summary at interval"}
			},
			flags: Flags{
				Files:    []string{"test1.log", "test2.log"},
				Interval: 100 * time.Millisecond,
			},
			rawRecords: [][]byte{
				[]byte(`{"time":"2025-08-14T02:07:12.680651416Z","host":"example.com","status_code":200,"duration":0.1}`),
				[]byte(`{"time":"2025-08-14T02:07:13.680651416Z","host":"example.com","status_code":404,"duration":0.2}`),
			},
			errChInput:    []error{},
			wantOutput:    []string{"Summary at interval"},
			wantErrOutput: []string{},
			wantReturn:    true,
			setupTimeout:  300 * time.Millisecond,
		},
		{
			name: "handles malformed records",
			setupSummarizer: func() *mockSummarizer {
				return &mockSummarizer{formatStr: "Summary with errors"}
			},
			flags: Flags{
				Files:    []string{"test1.log"},
				Interval: 200 * time.Millisecond,
			},
			rawRecords: [][]byte{
				[]byte(`invalid json`),
				[]byte(`{"time":"2025-08-14T02:07:12.680651416Z","host":"example.com","status_code":200,"duration":0.1}`),
				[]byte(`{"host":"incomplete.com"}`), // missing required fields
			},
			errChInput:    []error{},
			wantOutput:    []string{"Summary with errors", "Missing field or malformed log: 2"},
			wantErrOutput: []string{},
			wantReturn:    true,
			setupTimeout:  300 * time.Millisecond,
		},
		{
			name: "handles file errors and returns false",
			setupSummarizer: func() *mockSummarizer {
				return &mockSummarizer{formatStr: "Should not appear"}
			},
			flags: Flags{
				Files:    []string{"test1.log", "test2.log"},
				Interval: time.Second,
			},
			rawRecords: [][]byte{},
			errChInput: []error{
				errors.New("[test1.log] error tailing: file not found"),
				errors.New("[test2.log] error tailing: permission denied"),
			},
			wantOutput: []string{"File error summary:"},
			wantErrOutput: []string{
				"[test1.log] error tailing: file not found",
				"[test2.log] error tailing: permission denied",
			},
			wantReturn:   false,
			setupTimeout: 200 * time.Millisecond,
		},
		{
			name: "closes when rawRecords channel is closed",
			setupSummarizer: func() *mockSummarizer {
				return &mockSummarizer{formatStr: "Final summary"}
			},
			flags: Flags{
				Files:    []string{"test1.log"},
				Interval: time.Second, // Long interval, shouldn't trigger
			},
			rawRecords: [][]byte{
				[]byte(`{"time":"2025-08-14T02:07:12.680651416Z","host":"example.com","status_code":200,"duration":0.1}`),
			},
			errChInput:    []error{},
			wantOutput:    []string{"Final summary"},
			wantErrOutput: []string{},
			wantReturn:    true,
			setupTimeout:  200 * time.Millisecond,
		},
		{
			name: "periodic ticker triggers output",
			setupSummarizer: func() *mockSummarizer {
				return &mockSummarizer{formatStr: "Ticker summary"}
			},
			flags: Flags{
				Files:    []string{"test1.log"},
				Interval: 50 * time.Millisecond, // Short interval for testing
			},
			rawRecords: [][]byte{
				[]byte(`{"time":"2025-08-14T02:07:12.680651416Z","host":"example.com","status_code":200,"duration":0.1}`),
			},
			errChInput:    []error{},
			wantOutput:    []string{"Ticker summary"}, // Should see at least one ticker output
			wantErrOutput: []string{},
			wantReturn:    true,
			setupTimeout:  120 * time.Millisecond, // Enough for at least one tick
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			summarizer := tt.setupSummarizer()
			rawRecords := make(chan []byte, len(tt.rawRecords))
			errCh := make(chan error, len(tt.errChInput))
			var outputBuf, errBuf bytes.Buffer

			// Pre-populate channels
			for _, record := range tt.rawRecords {
				rawRecords <- record
			}
			for _, err := range tt.errChInput {
				errCh <- err
			}

			// Close channels based on test scenario
			if tt.wantReturn { // Expecting normal return (rawRecords closed)
				close(rawRecords)
			}
			close(errCh)

			// Run the function with timeout
			done := make(chan bool, 1)
			var result bool
			go func() {
				result = aggregateAndPrintSummaries(summarizer, tt.flags, rawRecords, errCh, &outputBuf, &errBuf)
				done <- true
			}()

			// Wait for completion or timeout
			select {
			case <-done:
				// Function completed
			case <-time.After(tt.setupTimeout):
				t.Fatalf("Test timed out after %v", tt.setupTimeout)
			}

			// Verify return value
			if result != tt.wantReturn {
				t.Errorf("aggregateAndPrintSummaries() returned %v, want %v", result, tt.wantReturn)
			}

			// Verify output
			outputStr := outputBuf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Output missing expected string %q. Got: %s", want, outputStr)
				}
			}

			// Verify error output
			errStr := errBuf.String()
			for _, want := range tt.wantErrOutput {
				if !strings.Contains(errStr, want) {
					t.Errorf("Error output missing expected string %q. Got: %s", want, errStr)
				}
			}

			// Verify summarizer was used correctly
			if len(tt.rawRecords) > 0 && len(summarizer.records) == 0 && !strings.Contains(outputStr, "Missing field or malformed log") {
				t.Error("Expected records to be processed by summarizer")
			}
		})
	}
}

