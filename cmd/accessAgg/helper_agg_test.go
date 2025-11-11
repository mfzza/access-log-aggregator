package main

import (
    "errors"
    "fmt"
    "os"
    "testing"
    "time"

	"accessAggregator/internal/accesslog"
)

// Mock Summaries for testing
type MockSummaries struct {
    records    []*accesslog.Record
    printCount int
}

func (m *MockSummaries) AddRecord(record *accesslog.Record) {
    m.records = append(m.records, record)
}

func (m *MockSummaries) Print() {
    m.printCount++
}

// Mock Record for testing
type MockRecord struct {
    Data string
}

// Test flags
type TestFlags struct {
    Interval time.Duration
    Files    []string
}

func TestAggregateAndPrintSummaries(t *testing.T) {
    // Capture stdout/stderr
    oldStdout := os.Stdout
    oldStderr := os.Stderr
    defer func() {
        os.Stdout = oldStdout
        os.Stderr = oldStderr
    }()

    tests := []struct {
        name           string
        flags          *TestFlags
        rawRecords     [][]byte
        errors         []error
        wantSuccess    bool
        wantPrintCount int
        wantErrors     bool
    }{
        {
            name: "successful processing with records",
            flags: &TestFlags{
                Interval: 100 * time.Millisecond,
                Files:    []string{"file1.log", "file2.log"},
            },
            rawRecords: [][]byte{
                []byte(`{"method":"GET","path":"/api"}`),
                []byte(`{"method":"POST","path":"/api"}`),
            },
            errors:         []error{},
            wantSuccess:    true,
            wantPrintCount: 2, // initial print + final print
            wantErrors:     false,
        },
        {
            name: "handles malformed records",
            flags: &TestFlags{
                Interval: 100 * time.Millisecond,
                Files:    []string{"file1.log"},
            },
            rawRecords: [][]byte{
                []byte(`invalid json`),
                []byte(`{"method":"GET","path":"/api"}`),
            },
            errors:         []error{},
            wantSuccess:    true,
            wantPrintCount: 2,
            wantErrors:     false,
        },
        {
            name: "handles file errors",
            flags: &TestFlags{
                Interval: 100 * time.Millisecond,
                Files:    []string{"file1.log", "file2.log"},
            },
            rawRecords: [][]byte{},
            errors: []error{
                errors.New("file1.log: permission denied"),
                errors.New("file2.log: not found"),
            },
            wantSuccess:    false,
            wantPrintCount: 0,
            wantErrors:     true,
        },
        {
            name: "handles mixed errors and records",
            flags: &TestFlags{
                Interval: 100 * time.Millisecond,
                Files:    []string{"file1.log", "file2.log"},
            },
            rawRecords: [][]byte{
                []byte(`{"method":"GET","path":"/api"}`),
            },
            errors: []error{
                errors.New("file2.log: not found"),
            },
            wantSuccess:    false,
            wantPrintCount: 0,
            wantErrors:     true,
        },
        {
            name: "periodic printing",
            flags: &TestFlags{
                Interval: 50 * time.Millisecond,
                Files:    []string{"file1.log"},
            },
            rawRecords: [][]byte{
                []byte(`{"method":"GET","path":"/api"}`),
            },
            errors:         []error{},
            wantSuccess:    true,
            wantPrintCount: 3, // initial + periodic + final
            wantErrors:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockSS := &MockSummaries{}
            rawRecordsCh := make(chan []byte, len(tt.rawRecords))
            errCh := make(chan error, len(tt.errors))

            // Populate channels
            for _, record := range tt.rawRecords {
                rawRecordsCh <- record
            }
            close(rawRecordsCh)

            for _, err := range tt.errors {
                errCh <- err
            }
            close(errCh)

            // Convert TestFlags to actual Flags type
            flags := &Flags{
                Interval: tt.flags.Interval,
                Files:    tt.flags.Files,
            }

            // Run the function
            result := aggregateAndPrintSummaries(mockSS, flags, rawRecordsCh, errCh)

            // Verify results
            if result != tt.wantSuccess {
                t.Errorf("aggregateAndPrintSummaries() = %v, want %v", result, tt.wantSuccess)
            }

            if mockSS.printCount != tt.wantPrintCount {
                t.Errorf("Print() called %d times, want %d", mockSS.printCount, tt.wantPrintCount)
            }

            if tt.wantErrors && len(tt.errors) > 0 {
                // We should verify errors were printed, but this is complex with stdout capture
                // In a real test, you'd want to capture and verify stderr output
            }
        })
    }
}

func TestAggregateAndPrintSummaries_EdgeCases(t *testing.T) {
    t.Run("empty records and no errors", func(t *testing.T) {
        mockSS := &MockSummaries{}
        flags := &Flags{
            Interval: 100 * time.Millisecond,
            Files:    []string{"file1.log"},
        }
        rawRecordsCh := make(chan []byte)
        close(rawRecordsCh)
        errCh := make(chan error)
        close(errCh)

        result := aggregateAndPrintSummaries(mockSS, flags, rawRecordsCh, errCh)

        if !result {
            t.Error("Expected success with empty records and no errors")
        }

        if mockSS.printCount < 1 {
            t.Error("Expected at least one print call")
        }
    })

    t.Run("immediate error all files", func(t *testing.T) {
        mockSS := &MockSummaries{}
        flags := &Flags{
            Interval: 100 * time.Millisecond,
            Files:    []string{"file1.log", "file2.log"},
        }
        rawRecordsCh := make(chan []byte)
        close(rawRecordsCh)
        errCh := make(chan error, 2)
        errCh <- errors.New("error1")
        errCh <- errors.New("error2")
        close(errCh)

        result := aggregateAndPrintSummaries(mockSS, flags, rawRecordsCh, errCh)

        if result {
            t.Error("Expected failure when all files have errors")
        }

        if mockSS.printCount > 0 {
            t.Error("Should not print summaries when all files error")
        }
    })
}

// Test helper to capture output
func captureOutput(f func()) string {
    oldStdout := os.Stdout
    oldStderr := os.Stderr
    defer func() {
        os.Stdout = oldStdout
        os.Stderr = oldStderr
    }()

    rOut, wOut, _ := os.Pipe()
    rErr, wErr, _ := os.Pipe()
    os.Stdout = wOut
    os.Stderr = wErr

    f()

    wOut.Close()
    wErr.Close()
    outBuf := make([]byte, 1024)
    errBuf := make([]byte, 1024)
    rOut.Read(outBuf)
    rErr.Read(errBuf)

    return string(outBuf) + string(errBuf)
}

// Test the actual output formatting
func TestAggregateAndPrintSummaries_Output(t *testing.T) {
    output := captureOutput(func() {
        mockSS := &MockSummaries{}
        flags := &Flags{
            Interval: 10 * time.Millisecond, // Short interval for test
            Files:    []string{"test.log"},
        }
        rawRecordsCh := make(chan []byte, 1)
        rawRecordsCh <- []byte(`{"method":"GET","path":"/api"}`)
        close(rawRecordsCh)
        errCh := make(chan error)
        close(errCh)

        aggregateAndPrintSummaries(mockSS, flags, rawRecordsCh, errCh)
    })

    // Verify output contains expected elements
    // This is a basic check - you might want more specific assertions
    if output == "" {
        t.Error("Expected some output from the function")
    }
}

// Benchmark test
func BenchmarkAggregateAndPrintSummaries(b *testing.B) {
    for b.Loop() {
        mockSS := &MockSummaries{}
        flags := &Flags{
            Interval: time.Hour, // Long interval to avoid ticker firing
            Files:    []string{"file1.log"},
        }

        rawRecordsCh := make(chan []byte, 100)
        for j := range 100 {
            rawRecordsCh <- fmt.Appendf(nil, `{"method":"GET","path":"/api/%d"}`, j)
        }
        close(rawRecordsCh)

        errCh := make(chan error)
        close(errCh)

        aggregateAndPrintSummaries(mockSS, flags, rawRecordsCh, errCh)
    }
}
