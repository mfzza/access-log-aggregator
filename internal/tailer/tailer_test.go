package tailer

import (
	"io"
	"os"
	"testing"
	"time"
)

func TestTailFile_GetRawRecord(t *testing.T) {
	mockFile := NewMockFile([]byte("line1\nline2\nline3\n"), "test.log")
	tailer, err := NewTailFile("test.log", mockFile, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create tailer: %v", err)
	}
	defer tailer.Close()

	// Test reading lines
	line, err := tailer.GetRawRecord()
	if err != nil {
		t.Fatalf("Failed to read line: %v", err)
	}
	if string(line) != "line1\n" {
		t.Errorf("Expected 'line1\\n', got '%s'", string(line))
	}

	line, err = tailer.GetRawRecord()
	if err != nil {
		t.Fatalf("Failed to read line: %v", err)
	}
	if string(line) != "line2\n" {
		t.Errorf("Expected 'line2\\n', got '%s'", string(line))
	}
}

func TestTailFile_EOF_Retry(t *testing.T) {
	mockFile := NewMockFile([]byte("line1\n"), "test.log")
	tailer, err := NewTailFile("test.log", mockFile, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create tailer: %v", err)
	}
	defer tailer.Close()

	// Read the only line
	line, err := tailer.GetRawRecord()
	if err != nil {
		t.Fatalf("Failed to read line: %v", err)
	}
	if string(line) != "line1\n" {
		t.Errorf("Expected 'line1\\n', got '%s'", string(line))
	}

	// Next read should return EOF but not break
	start := time.Now()
	line, err = tailer.GetRawRecord()
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}

	// Should have delayed
	if time.Since(start) < 5*time.Millisecond {
		t.Error("Should have delayed on EOF")
	}
}

func TestTailFile_TruncationDetection(t *testing.T) {
	callCount := 0
	mockFile := NewMockFile([]byte("original content\n"), "test.log")

	// Simulate file truncation
	mockFile.NewStatFunc = func(path string) (os.FileInfo, error) {
		callCount++
		if callCount == 1 {
			// First call - normal size
			return &MockFileInfo{
				NameVal: path,
				SizeVal: 18,
			}, nil
		}
		// Second call - truncated
		return &MockFileInfo{
			NameVal: path,
			SizeVal: 0,
		}, nil
	}

	tailer, err := NewTailFile("test.log", mockFile, time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create tailer: %v", err)
	}
	defer tailer.Close()

	// Simulate EOF to trigger rotation check
	mockFile.ReadPos = len(mockFile.Content)
	_, err = tailer.GetRawRecord()

	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}

	// After truncation detection, should seek to start
	if mockFile.SeekPos != 0 {
		t.Errorf("Expected seek to 0 after truncation, got %d", mockFile.SeekPos)
	}
}

func TestTailFile_RenameDetection(t *testing.T) {
	callCount := 0
	mockFile := NewMockFile([]byte("old content\n"), "test.log")

	mockFile.NewStatFunc = func(path string) (os.FileInfo, error) {
		callCount++
		if callCount == 1 {
			// Same file initially
			return &MockFileInfo{
				NameVal: "test.log",
				SizeVal: 12,
			}, nil
		}
		// Different file (renamed)
		return &MockFileInfo{
			NameVal: "test.log.1",
			SizeVal: 12,
		}, nil
	}

	tailer, err := NewTailFile("test.log", mockFile, time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create tailer: %v", err)
	}
	defer tailer.Close()

	// First EOF - should detect rotation but not reopen yet
	mockFile.ReadPos = len(mockFile.Content)
	_, err = tailer.GetRawRecord()
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}

	if !tailer.rotated {
		t.Error("Should have detected file rotation")
	}

	// Second EOF - should reopen the file
	mockFile.OpenFunc = func(name string) (FileSrc, error) {
		return NewMockFile([]byte("new content\n"), name), nil
	}

	_, err = tailer.GetRawRecord()
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}

	if !tailer.rotated {
		t.Error("Should have cleared rotated flag after reopening")
	}
}

func TestTailFile_TruncationSeeksToStart(t *testing.T) {
	callCount := 0
	mockFile := NewMockFile([]byte("original content\nmore content\n"), "test.log")

	mockFile.NewStatFunc = func(path string) (os.FileInfo, error) {
		callCount++
		if callCount == 1 {
			// First call - normal size (18 bytes)
			return &MockFileInfo{SizeVal: 18}, nil
		}
		// Second call - truncated to 0 bytes
		return &MockFileInfo{SizeVal: 0}, nil
	}

	// Track seek calls
	mockFile.SeekFunc = func(offset int64, whence int) (int64, error) {
		if offset != 0 || whence != io.SeekStart {
			t.Errorf("Expected Seek(0, io.SeekStart), got Seek(%d, %d)", offset, whence)
		}
		return mockFile.Seek(offset, whence)
	}

	tailer, err := NewTailFile("test.log", mockFile, time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create tailer: %v", err)
	}
	defer tailer.Close()

	// Read first line to establish position
	line, err := tailer.GetRawRecord()
	if err != nil {
		t.Fatalf("Failed to read line: %v", err)
	}
	if string(line) != "original content\n" {
		t.Errorf("Expected 'original content\\n', got '%s'", string(line))
	}

	// Simulate EOF to trigger rotation check with truncation
	mockFile.ReadPos = len(mockFile.Content)
	_, err = tailer.GetRawRecord()

	// Verify reader was reset
	if mockFile.SeekPos != 0 {
		t.Errorf("Expected file position to be 0 after truncation, got %d", mockFile.SeekPos)
	}
}

func TestTailFile_Close(t *testing.T) {
	mockFile := NewMockFile([]byte("content\n"), "test.log")
	tailer, err := NewTailFile("test.log", mockFile, time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create tailer: %v", err)
	}

	err = tailer.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if !mockFile.Closed {
		t.Error("Underlying file should be closed")
	}
}

func TestTailFile_ErrorHandling(t *testing.T) {
	// Test with file that errors on NewStat
	mockFile := NewMockFile([]byte("content\n"), "test.log")
	mockFile.NewStatFunc = func(path string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	tailer, err := NewTailFile("test.log", mockFile, time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create tailer: %v", err)
	}
	defer tailer.Close()

	// Should not break on Stat errors during rotation check
	mockFile.ReadPos = len(mockFile.Content)
	_, err = tailer.GetRawRecord()
	// Should get EOF, not the Stat error
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
}

func TestWhichRotation(t *testing.T) {
	file1 := &MockFileInfo{NameVal: "file1", SizeVal: 100}
	file2 := &MockFileInfo{NameVal: "file2", SizeVal: 50}

	originalSameFile := sameFile
	defer func() { sameFile = originalSameFile }()

	sameFile = func(fi1, fi2 os.FileInfo) bool { return true }
	result := whichRotation(file1, file1)
	if result != same {
		t.Errorf("Expected same, got %v", result)
	}

	// Test truncation
	smallerFile := &MockFileInfo{NameVal: "file1", SizeVal: 50}
	result = whichRotation(file1, smallerFile)
	if result != truncated {
		t.Errorf("Expected truncated, got %v", result)
	}

	// Test rename
	sameFile = func(fi1, fi2 os.FileInfo) bool { return false }
	result = whichRotation(file1, file2)
	if result != renamed {
		t.Errorf("Expected renamed, got %v", result)
	}
}
