// tailer_test.go
package tailer

import (
	"errors"
	"io"
	"os"
	"testing"
)

func TestNewTailFile(t *testing.T) {
	tests := []struct {
		name      string
		setupFS   func() *mockFileSystem
		fromStart bool
		wantErr   bool
	}{
		{
			name: "successful creation from start",
			setupFS: func() *mockFileSystem {
				fs := newMockFileSystem()
				fs.files["test.log"] = newMockFile([]byte("line1\nline2\n"))
				return fs
			},
			fromStart: true,
			wantErr:   false,
		},
		{
			name: "successful creation from end",
			setupFS: func() *mockFileSystem {
				fs := newMockFileSystem()
				fs.files["test.log"] = newMockFile([]byte("line1\nline2\n"))
				return fs
			},
			fromStart: false,
			wantErr:   false,
		},
		{
			name: "file not found",
			setupFS: func() *mockFileSystem {
				return newMockFileSystem() // empty, no files
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.setupFS()
			tailer, err := NewTailFile("test.log", fs, tt.fromStart)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewTailFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tailer == nil {
				t.Error("NewTailFile() returned nil tailer without error")
			}

			if tailer != nil {
				tailer.Close()
			}
		})
	}
}

func TestGetRawRecord(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		wantLines   []string
		wantEOF     bool
	}{
		{
			name:        "read multiple lines",
			fileContent: "line1\nline2\nline3\n",
			wantLines:   []string{"line1\n", "line2\n", "line3\n"},
			wantEOF:     true,
		},
		{
			name:        "read incomplete line",
			fileContent: "incomplete line",
			wantLines:   []string{},
			wantEOF:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()
			fs.files["test.log"] = newMockFile([]byte(tt.fileContent))

			tailer, err := NewTailFile("test.log", fs, true)
			if err != nil {
				t.Fatalf("Failed to create tailer: %v", err)
			}
			defer tailer.Close()

			for _, wantLine := range tt.wantLines {
				line, err := tailer.GetRawRecord()
				if err != nil {
					t.Errorf("GetRawRecord() unexpected error: %v", err)
				}
				if string(line) != wantLine {
					t.Errorf("GetRawRecord() = %q, want %q", string(line), wantLine)
				}
			}

			// Test EOF behavior
			if tt.wantEOF {
				line, err := tailer.GetRawRecord()
				if err != io.EOF {
					t.Errorf("GetRawRecord() error = %v, want EOF", err)
				}
				if line != nil {
					t.Errorf("GetRawRecord() line = %v, want nil on EOF", line)
				}
			}
		})
	}
}

func TestCheckRotation(t *testing.T) {
	tests := []struct {
		name           string
		initialContent string
		initialSize    int64
		newContent     string
		newSize        int64
		sameFile       bool
		wantReset      bool
		wantReopen     bool
	}{
		{
			name:           "file truncated",
			initialContent: "line1\nline2\nline3\n",
			initialSize:    18,
			newContent:     "new1\n",
			newSize:        5,
			sameFile:       true,
			wantReset:      true,
			wantReopen:     false,
		},
		{
			name:           "file renamed and new created",
			initialContent: "line1\nline2\n",
			initialSize:    12,
			newContent:     "new1\nnew2\n",
			newSize:        10,
			sameFile:       false,
			wantReset:      false,
			wantReopen:     true,
		},
		{
			name:           "no rotation",
			initialContent: "line1\nline2\n",
			initialSize:    12,
			newContent:     "line1\nline2\nline3\n",
			newSize:        18,
			sameFile:       true,
			wantReset:      false,
			wantReopen:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFileSystem()

			// Setup initial file
			initialFile := newMockFile([]byte(tt.initialContent))
			initialFile.statFunc = func() (os.FileInfo, error) {
				return &mockFileInfo{name: "test.log", size: tt.initialSize}, nil
			}
			fs.files["test.log"] = initialFile

			tailer, err := NewTailFile("test.log", fs, true)
			if err != nil {
				t.Fatalf("Failed to create tailer: %v", err)
			}

			// Setup new file state
			newFile := newMockFile([]byte(tt.newContent))
			newFile.statFunc = func() (os.FileInfo, error) {
				return &mockFileInfo{name: "test.log", size: tt.newSize}, nil
			}

			// Mock the file system stat to return new file info
			fs.statFunc = func(name string) (os.FileInfo, error) {
				return newFile.Stat()
			}

			// Mock os.SameFile for this test
			originalSameFile := sameFile
			sameFile = func(fi1, fi2 os.FileInfo) bool { return tt.sameFile }
			defer func() { sameFile = originalSameFile }()

			// Mock open for rotation case
			if tt.wantReopen {
				fs.openFunc = func(name string) (file, error) {
					return newFile, nil
				}
			}

			err = tailer.checkRotation()
			if err != nil {
				t.Errorf("checkRotation() unexpected error: %v", err)
			}

			// Verify results based on expectations
			if tt.wantReset {
				// Should have seeked to start
				pos, _ := initialFile.Seek(0, io.SeekCurrent)
				if pos != 0 {
					t.Error("File should be reset to start after truncation")
				}
			}

			if tt.wantReopen {
				initialStat, _  := initialFile.Stat()
				if sameFile(tailer.fstat, initialStat) {
					t.Error("File should be reopened after rotation")
				}
			}

			tailer.Close()
		})
	}
}

func TestClose(t *testing.T) {
	fs := newMockFileSystem()
	mockFile := newMockFile([]byte("test content"))
	fs.files["test.log"] = mockFile

	tailer, err := NewTailFile("test.log", fs, true)
	if err != nil {
		t.Fatalf("Failed to create tailer: %v", err)
	}

	err = tailer.Close()
	if err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}

	if !mockFile.closed {
		t.Error("Underlying file was not closed")
	}
}

func TestErrorScenarios(t *testing.T) {
	t.Run("stat error during rotation check", func(t *testing.T) {
		fs := newMockFileSystem()
		fs.files["test.log"] = newMockFile([]byte("content"))

		tailer, err := NewTailFile("test.log", fs, true)
		if err != nil {
			t.Fatalf("Failed to create tailer: %v", err)
		}
		defer tailer.Close()

		// Mock stat to return error
		fs.statFunc = func(name string) (os.FileInfo, error) {
			return nil, errors.New("stat error")
		}

		err = tailer.checkRotation()
		if err != nil {
			t.Errorf("checkRotation() should handle stat errors gracefully, got: %v", err)
		}
	})

	t.Run("reopen error during rotation", func(t *testing.T) {
		fs := newMockFileSystem()
		initialFile := newMockFile([]byte("content"))
		fs.files["test.log"] = initialFile

		tailer, err := NewTailFile("test.log", fs, true)
		if err != nil {
			t.Fatalf("Failed to create tailer: %v", err)
		}
		defer tailer.Close()

		// Setup rotation scenario
		fs.statFunc = func(name string) (os.FileInfo, error) {
			return &mockFileInfo{name: "test.log", size: 100}, nil
		}

		originalSameFile := sameFile
		sameFile = func(fi1, fi2 os.FileInfo) bool { return false }
		defer func() { sameFile = originalSameFile }()

		// Mock open to return error
		fs.openFunc = func(name string) (file, error) {
			return nil, errors.New("open failed")
		}

		// First call marks rotated
		tailer.checkRotation()

		// Second call should try to reopen and fail
		err = tailer.checkRotation()
		if err == nil {
			t.Error("Expected error when reopening fails")
		}
	})
}
