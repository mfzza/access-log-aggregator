//go:build integration

package integration_test

import (
	"accessAggregator/internal/app"
	"accessAggregator/internal/config"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// test the complete workflow from file reading to summary output
func TestEndToEnd(t *testing.T) {
	tests := []struct {
		name        string
		logFiles    map[string]string // [filename]content
		fromStart   bool
		interval    time.Duration
		expectHosts []string
		waitTime    time.Duration
	}{
		{
			name:      "single file with multiple hosts",
			fromStart: true,
			interval:  500 * time.Millisecond,
			logFiles: map[string]string{
				"test1.log": `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
{"time":"2025-08-14T02:07:13.680651416Z","host":"github.com","status_code":404,"duration":0.150}
{"time":"2025-08-14T02:07:14.680651416Z","host":"chatgpt.com","status_code":201,"duration":0.300}
`,
			},
			expectHosts: []string{"chatgpt.com", "github.com"},
			waitTime:    600 * time.Millisecond,
		},
		{
			name:      "multiple files",
			fromStart: true,
			interval:  500 * time.Millisecond,
			logFiles: map[string]string{
				"test1.log": `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
`,
				"test2.log": `{"time":"2025-08-14T02:07:13.680651416Z","host":"github.com","status_code":200,"duration":0.150}
`,
			},
			expectHosts: []string{"chatgpt.com", "github.com"},
			waitTime:    600 * time.Millisecond,
		},
		{
			name:      "with malformed records",
			fromStart: true,
			interval:  500 * time.Millisecond,
			logFiles: map[string]string{
				"test1.log": `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
{"invalid json
{"time":"2025-08-14T02:07:14.680651416Z","host":"github.com","status_code":200,"duration":0.300}
`,
			},
			expectHosts: []string{"chatgpt.com", "github.com"},
			waitTime:    600 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			var filePaths []string
			for filename, content := range tt.logFiles {
				fpath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(fpath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				filePaths = append(filePaths, fpath)
			}

			flags := config.Flags{
				Files:     filePaths,
				FromStart: tt.fromStart,
				Interval:  tt.interval,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			var out bytes.Buffer

			done := make(chan error, 1)
			go func() {
				done <- app.Run(ctx, flags, &out, io.Discard)
			}()

			// wait for processing
			time.Sleep(tt.waitTime)
			cancel()

			// wait for completion
			select {
			case err := <-done:
				if err != nil && !strings.Contains(err.Error(), "context") {
					t.Errorf("Run() error = %v", err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("Test timeout")
			}

			for _, host := range tt.expectHosts {
				if !strings.Contains(out.String(), host) {
					t.Fatal(host, "not found in output")
				}
			}
		})
	}
}

func TestNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does-not-exist.log")

	flags := config.Flags{
		Files:     []string{nonExistentFile},
		FromStart: true,
		Interval:  500 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, flags, &out, &errOut)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Test timeout")
	}

	if !strings.Contains(errOut.String(), "does-not-exist.log") {
		t.Error("Expected filename in error output")
	}
	if !strings.Contains(errOut.String(), "error:") {
		t.Error("Expected error message")
	}
}

// test with malformed record
func TestMalformedRecord(t *testing.T) {
	tests := []struct {
		name            string
		logContent      string
		expectMalformed bool
		expectedHosts   []string
		unexpectedHosts []string
	}{
		{
			name: "mixed valid and malformed JSON",
			logContent: `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
{"time":"2025-08-14T02:07:13.680651416Z","host":"github.com","status_code":404,"duration":0.150}
{"invalid json without closing brace
{"time":"2025-08-14T02:07:14.680651416Z","host":"openai.com","status_code":201,"duration":0.300}
{this is also bad json}
{"time":"2025-08-14T02:07:15.680651416Z","host":"example.com","status_code":500,"duration":0.100}
`,
			expectMalformed: true,
			expectedHosts:   []string{"chatgpt.com", "github.com", "openai.com", "example.com"},
		},
		{
			name: "missing required fields",
			logContent: `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
{"host":"github.com","status_code":404,"duration":0.150}
{"time":"2025-08-14T02:07:13.680651416Z","status_code":200,"duration":0.300}
{"time":"2025-08-14T02:07:14.680651416Z","host":"openai.com","duration":0.100}
{"time":"2025-08-14T02:07:15.680651416Z","host":"example.com","status_code":201}
{"time":"2025-08-14T02:07:16.680651416Z","host":"valid.com","status_code":200,"duration":0.150}
`,
			expectMalformed: true,
			expectedHosts:   []string{"chatgpt.com", "valid.com"},
			unexpectedHosts: []string{"github.com", "openai.com", "example.com"},
		},
		{
			name: "zero values in required fields",
			logContent: `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
{"time":"2025-08-14T02:07:13.680651416Z","host":"github.com","status_code":0,"duration":0.150}
{"time":"2025-08-14T02:07:14.680651416Z","host":"openai.com","status_code":200,"duration":0}
`,
			expectMalformed: true,
			expectedHosts:   []string{"chatgpt.com"},
			unexpectedHosts: []string{"github.com", "openai.com"},
		},
		{
			name: "all malformed",
			logContent: `invalid line 1
{bad json
{"missing":"required_fields"}
completely broken
`,
			expectMalformed: true,
			expectedHosts:   []string{},
		},
		{
			name: "all valid",
			logContent: `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
{"time":"2025-08-14T02:07:13.680651416Z","host":"github.com","status_code":404,"duration":0.150}
{"time":"2025-08-14T02:07:14.680651416Z","host":"openai.com","status_code":201,"duration":0.300}
`,
			expectMalformed: false,
			expectedHosts:   []string{"chatgpt.com", "github.com", "openai.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "test.log")

			if err := os.WriteFile(logFile, []byte(tt.logContent), 0644); err != nil {
				t.Fatalf("Failed to create log file: %v", err)
			}

			flags := config.Flags{
				Files:     []string{logFile},
				FromStart: true,
				Interval:  500 * time.Millisecond,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			var out bytes.Buffer
			done := make(chan error, 1)
			go func() {
				done <- app.Run(ctx, flags, &out, io.Discard)
			}()

			time.Sleep(600 * time.Millisecond)
			cancel()

			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("Test timeout")
			}

			output := out.String()

			// check malformed message
			hasMalformed := strings.Contains(output, "missing field or malformed log:")
			if tt.expectMalformed && !hasMalformed {
				t.Error("Expected malformed log message, but not found")
			}
			if !tt.expectMalformed && hasMalformed {
				t.Error("Unexpected malformed log message found")
			}

			// verify expected hosts appear
			for _, host := range tt.expectedHosts {
				if !strings.Contains(output, host) {
					t.Errorf("Expected host %q not found in output", host)
				}
			}

			// verify unexpected hosts don't appear
			for _, host := range tt.unexpectedHosts {
				if strings.Contains(output, host) {
					t.Errorf("Unexpected host %q found in output (should have been filtered)", host)
				}
			}

			// verify final summary was printed
			if !strings.Contains(output, "Printing final summary:") {
				t.Error("Final summary not printed")
			}
		})
	}
}

// test log rotation scenarios
func TestFileRotationTruncated(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "rotating.log")

	initialContent := `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
`
	if err := os.WriteFile(logFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	flags := config.Flags{
		Files:     []string{logFile},
		FromStart: true,
		Interval:  200 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var out bytes.Buffer

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, flags, &out, io.Discard)
	}()

	// wait a bit, then append more content
	time.Sleep(300 * time.Millisecond)
	newContent := `{"time":"2025-08-14T02:07:13.680651416Z","host":"github.com","status_code":200,"duration":0.150}
`
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file for append: %v", err)
	}
	f.WriteString(newContent)

	// truncate
	time.Sleep(300 * time.Millisecond)
	f.Truncate(0)
	f.Seek(0, io.SeekStart)

	time.Sleep(300 * time.Millisecond)
	truncatedContent := `{"time":"2025-08-14T02:07:14.680651416Z","host":"truncated.com","status_code":200,"duration":0.100}
`
	f.WriteString(truncatedContent)

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success - processed rotation
	case <-time.After(4 * time.Second):
		t.Fatal("Test timeout")
	}

	// check new host
	expectHosts := []string{"chatgpt.com", "github.com", "truncated.com"}
	for _, host := range expectHosts {
		if !strings.Contains(out.String(), host) {
			t.Fatal("new host from log rotation (truncated) file not found in output")
		}
	}
}

func TestFileRotationRenamed(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "rotating.log")

	initialContent := `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
`
	if err := os.WriteFile(logFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	flags := config.Flags{
		Files:     []string{logFile},
		FromStart: true,
		Interval:  200 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var out bytes.Buffer

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, flags, &out, io.Discard)
	}()

	// wait a bit, then append more content
	time.Sleep(300 * time.Millisecond)
	newContent := `{"time":"2025-08-14T02:07:13.680651416Z","host":"github.com","status_code":200,"duration":0.150}
`
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file for append: %v", err)
	}
	f.WriteString(newContent)

	// renamed
	backupFile := logFile + ".1"
	if err := os.Rename(logFile, backupFile); err != nil {
		t.Fatalf("Failed to rename file: %v", err)
	}

	renamedContent := `{"time":"2025-08-14T02:07:15.680651416Z","host":"renamed.com","status_code":200,"duration":0.200}
`
	if err := os.WriteFile(logFile, []byte(renamedContent), 0644); err != nil {
		t.Fatalf("Failed to create rotated file: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success - processed rotation
	case <-time.After(4 * time.Second):
		t.Fatal("Test timeout")
	}

	// check new host, also make sure it also proceed til eof when renamed
	expectHosts := []string{"chatgpt.com", "github.com", "renamed.com"}
	for _, host := range expectHosts {
		if !strings.Contains(out.String(), host) {
			t.Fatal("new host from log rotation (renamed) file not found in output")
		}
	}
}

// test multiple files being read concurrently
func TestConcurrentFileReading(t *testing.T) {
	tmpDir := t.TempDir()

	// create multiple log files
	numFiles := 100
	var filePaths []string
	for i := range numFiles {
		fpath := filepath.Join(tmpDir, "test"+strconv.Itoa(+i)+".log")
		content := ""
		for range 10 {
			content += `{"time":"2025-08-14T02:07:12.680651416Z","host":"host` + strconv.Itoa(i) + `.com","status_code":200,"duration":0.100}
`
		}
		if err := os.WriteFile(fpath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		filePaths = append(filePaths, fpath)
	}

	flags := config.Flags{
		Files:     filePaths,
		FromStart: true,
		Interval:  100 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var out bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, flags, &out, io.Discard)
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && !strings.Contains(err.Error(), "context") {
			t.Errorf("Run() error = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Test timeout")
	}

	for i := range numFiles {
		host := "host" + strconv.Itoa(i) + ".com"
		if !strings.Contains(out.String(), host) {
			t.Fatal("Missing " + host + " in output")
		}
	}
	// t.Log(out.String())
}

// test graceful shutdown
func TestGracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// create a file with many records
	numRecords := 100
	content := ""
	for range numRecords {
		content += `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
`
	}
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	flags := config.Flags{
		Files:     []string{logFile},
		FromStart: true,
		Interval:  1 * time.Second, // Long interval
	}

	ctx, cancel := context.WithCancel(context.Background())

	var out bytes.Buffer

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, flags, &out, io.Discard)
	}()

	// let it process some records first, then cancel context
	time.Sleep(50 * time.Millisecond)
	cancel()

	// should complete within reasonable time
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Graceful shutdown took too long")
	}

	output := out.String()

	// verify graceful shutdown:

	// 1. final summary was printed
	if !strings.Contains(output, "Printing final summary:") {
		t.Error("Final summary not printed - shutdown wasn't graceful")
	}

	// 2. host appears in output (data was processed)
	if !strings.Contains(output, "chatgpt.com") {
		t.Error("Expected host not found - data wasn't processed")
	}

	// 3. shutdown message was printed
	if !strings.Contains(output, "Gracefully shut down...") {
		t.Error("Graceful shutdown message not found")
	}

	// 4. all 100 records were processed (check the count in summary)
	// This verifies the channel was drained before exit
	if !strings.Contains(output, strconv.Itoa(numRecords)) {
		t.Error("Record not drained before exit")
	}
}
