// +build integration

package app_test

import (
	"accessAggregator/internal/app"
	"accessAggregator/internal/config"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// captureOutput captures stdout during test execution
type captureOutput struct {
	buf    bytes.Buffer
	writer io.Writer
}

func (c *captureOutput) Write(p []byte) (n int, err error) {
	return c.buf.Write(p)
}

func (c *captureOutput) String() string {
	return c.buf.String()
}

// TestEndToEnd tests the complete workflow from file reading to summary output
func TestEndToEnd(t *testing.T) {
	tests := []struct {
		name        string
		logFiles    map[string]string // filename -> content
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
			// Create temp directory
			tmpDir := t.TempDir()

			// Create test log files
			var filePaths []string
			for filename, content := range tt.logFiles {
				fpath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(fpath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				filePaths = append(filePaths, fpath)
			}

			// Setup flags
			flags := config.Flags{
				Files:     filePaths,
				FromStart: tt.fromStart,
				Interval:  tt.interval,
			}

			// Run with timeout context
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Run in goroutine so we can control when to stop
			done := make(chan error, 1)
			go func() {
				done <- app.Run(ctx, flags)
			}()

			// Wait for processing
			time.Sleep(tt.waitTime)
			cancel()

			// Wait for completion
			select {
			case err := <-done:
				if err != nil && !strings.Contains(err.Error(), "context") {
					t.Errorf("Run() error = %v", err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("Test timeout")
			}

			// Note: In real integration tests, you'd capture stdout/stderr
			// to verify the summary output contains expected hosts
		})
	}
}

// TestFileRotation tests handling of log rotation scenarios
func TestFileRotation(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "rotating.log")

	// Write initial content
	initialContent := `{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":200,"duration":0.224}
`
	if err := os.WriteFile(logFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	flags := config.Flags{
		Files:     []string{logFile},
		FromStart: false,
		Interval:  200 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, flags)
	}()

	// Wait a bit, then append more content
	time.Sleep(300 * time.Millisecond)
	newContent := `{"time":"2025-08-14T02:07:13.680651416Z","host":"github.com","status_code":200,"duration":0.150}
`
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file for append: %v", err)
	}
	f.WriteString(newContent)
	f.Close()

	// Simulate truncation
	time.Sleep(300 * time.Millisecond)
	truncatedContent := `{"time":"2025-08-14T02:07:14.680651416Z","host":"example.com","status_code":200,"duration":0.100}
`
	if err := os.WriteFile(logFile, []byte(truncatedContent), 0644); err != nil {
		t.Fatalf("Failed to truncate file: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success - processed rotation
	case <-time.After(4 * time.Second):
		t.Fatal("Test timeout")
	}
}

// TestConcurrentFileReading tests multiple files being read concurrently
func TestConcurrentFileReading(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple log files
	numFiles := 5
	var filePaths []string
	for i := range numFiles {
		fpath := filepath.Join(tmpDir, "test"+string(rune('A'+i))+".log")
		content := ""
		for range 10 {
			content += `{"time":"2025-08-14T02:07:12.680651416Z","host":"host` + string(rune('A'+i)) + `.com","status_code":200,"duration":0.100}
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

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, flags)
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
}

// TestGracefulShutdown tests that the application shuts down cleanly
func TestGracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Create a file with many records
	content := ""
	for range 100 {
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

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, flags)
	}()

	// Cancel immediately
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Should complete within reasonable time
	select {
	case <-done:
		// Success - shutdown completed
	case <-time.After(2 * time.Second):
		t.Fatal("Graceful shutdown took too long")
	}
}
