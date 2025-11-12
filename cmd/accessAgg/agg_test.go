package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"accessAggregator/internal/accesslog"
)

// Mock Summarizer for testing
type mockSummarizer struct {
	records     []accesslog.Record
	printCalled int
}

func (m *mockSummarizer) AddRecord(r *accesslog.Record) {
	m.records = append(m.records, *r)
}

func (m *mockSummarizer) Print() {
	m.printCalled++
}

// Test cases
func TestAggregateAndPrintSummaries(t *testing.T) {
	t.Run("successful processing with periodic printing", func(t *testing.T) {
		// Setup
		ms := &mockSummarizer{}
		flags := Flags{Interval: 100 * time.Millisecond}
		rawRecords := make(chan []byte, 10)
		errCh := make(chan error, 1)
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		// Test data
		validRecord := []byte(`{"time":"2025-10-21T22:48:42Z","host":"ahatgpt.com","status_code":156,"duration":0.861705397}`)
		rawRecords <- validRecord
		rawRecords <- validRecord

		// Run function in goroutine
		done := make(chan bool)
		go func() {
			result := aggregateAndPrintSummaries(ms, flags, rawRecords, errCh, out, errOut)
			done <- result
		}()

		// Wait a bit for processing
		time.Sleep(50 * time.Millisecond)
		close(rawRecords)

		// Verify
		result := <-done
		if !result {
			t.Error("Expected true result, got false")
		}
		if ms.printCalled == 0 {
			t.Error("Print should have been called")
		}
		if len(ms.records) != 2 {
			t.Errorf("Expected 2 records, got %d", len(ms.records))
		}
	})

	t.Run("handles malformed records", func(t *testing.T) {
		ms := &mockSummarizer{}
		flags := Flags{Interval: time.Hour} // Long interval to avoid ticker
		rawRecords := make(chan []byte, 10)
		errCh := make(chan error, 1)
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		// Send malformed record
		rawRecords <- []byte("invalid json")
		close(rawRecords)

		result := aggregateAndPrintSummaries(ms, flags, rawRecords, errCh, out, errOut)

		if !result {
			t.Error("Expected true result for malformed records")
		}
		if len(ms.records) != 0 {
			t.Error("No records should be added for malformed data")
		}
		// Check that broken record count is printed
		output := out.String()
		if !strings.Contains(output, "Malformed log:") {
			t.Error("Should report malformed logs in output")
		}
	})

	t.Run("handles errors from error channel", func(t *testing.T) {
		ms := &mockSummarizer{}
		flags := Flags{
			Interval: time.Hour,
			Files:    []string{"file1", "file2"}, // 2 files
		}
		rawRecords := make(chan []byte, 10)
		errCh := make(chan error, 2)
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		// Send errors matching number of files
		errCh <- errors.New("file1 error")
		errCh <- errors.New("file2 error")
		close(errCh)

		result := aggregateAndPrintSummaries(ms, flags, rawRecords, errCh, out, errOut)

		if result {
			t.Error("Expected false result when all files have errors")
		}

		errorOutput := errOut.String()
		if !strings.Contains(errorOutput, "file1 error") || !strings.Contains(errorOutput, "file2 error") {
			t.Error("Should print all errors to error output")
		}
	})

	t.Run("periodic printing with ticker", func(t *testing.T) {
		ms := &mockSummarizer{}
		flags := Flags{Interval: 50 * time.Millisecond}
		rawRecords := make(chan []byte, 10)
		errCh := make(chan error, 1)
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		done := make(chan bool)
		go func() {
			result := aggregateAndPrintSummaries(ms, flags, rawRecords, errCh, out, errOut)
			done <- result
		}()

		// Wait for multiple ticker intervals
		time.Sleep(220 * time.Millisecond)
		close(rawRecords)

		result := <-done
		if !result {
			t.Error("Expected true result")
		}

		// Should have multiple print calls (initial + ticker)
		if ms.printCalled < 2 {
			t.Errorf("Expected at least 2 print calls, got %d", ms.printCalled)
		}
	})
}
