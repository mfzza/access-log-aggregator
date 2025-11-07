package fileutil_test

import (
	"accessAggregator/internal/fileutil"
	"bufio"
	"errors"
	"io"
	"os"
	"testing"
	"time"
)

type mockFile struct {
	data     []byte
	offset   int64
	stat     os.FileInfo
	failStat bool
	failSeek bool
}

func (m *mockFile) Read(p []byte) (int, error) {
	if m.offset >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n := copy(p, m.data[m.offset:])
	m.offset += int64(n)
	return n, nil
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	if m.failSeek {
		return 0, errors.New("seek failed")
	}
	switch whence {
	case io.SeekStart:
		m.offset = offset
	case io.SeekEnd:
		m.offset = int64(len(m.data)) + offset
	case io.SeekCurrent:
		m.offset += offset
	}
	return m.offset, nil
}

func (m *mockFile) Stat() (os.FileInfo, error) {
	if m.failStat {
		return nil, errors.New("stat failed")
	}
	return m.stat, nil
}

type mockStat struct{ size int64 }

func (m mockStat) Name() string       { return "mock.log" }
func (m mockStat) Size() int64        { return m.size }
func (m mockStat) Mode() os.FileMode  { return 0 }
func (m mockStat) ModTime() time.Time { return time.Now() }
func (m mockStat) IsDir() bool        { return false }
func (m mockStat) Sys() any           { return nil }

func TestNewTailFile(t *testing.T) {
	origOpenFile := fileutil.OpenFile // save original
	defer func() { fileutil.OpenFile = origOpenFile }()

	mock := &mockFile{
		data: []byte("line1\nline2\n"),
		stat: mockStat{size: 12},
	}

	fileutil.OpenFile = func(fpath string) (fileutil.FileSource, error) {
		if fpath == "missing.log" {
			return nil, errors.New("open failed")
		}
		return mock, nil
	}

	tests := []struct {
		name      string
		fpath     string
		fromStart bool
		wantErr   bool
	}{
		{"success from start", "ok.log", true, false},
		{"success from end", "ok.log", false, false},
		{"open error", "missing.log", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fileutil.NewTailFile(tt.fpath, tt.fromStart)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected non-nil TailFile")
			}
		})
	}
}

type stubTailFile struct {
	*fileutil.TailFile
	checkRotationFunc func() error
}

func (s *stubTailFile) checkRotation() error {
	if s.checkRotationFunc != nil {
		return s.checkRotationFunc()
	}
	return nil
}

func TestNextLine(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		readErr          error // simulate custom reader error
		checkRotationErr error
		expectLine       string
		expectErr        error
	}{
		{
			name:       "normal line read",
			input:      "hello world\n",
			expectLine: "hello world\n",
			expectErr:  nil,
		},
		{
			name:      "EOF without rotation error",
			input:     "",
			expectErr: io.EOF,
		},
		{
			name:             "EOF with rotation error",
			input:            "",
			checkRotationErr: errors.New("rotation fail"),
			expectErr:        errors.New("rotation fail"), // wrapped checkRotation error
		},
		{
			name:      "unexpected read error",
			input:     "",
			readErr:   errors.New("disk I/O fail"),
			expectErr: errors.New("disk I/O fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			reader := &faultyReader{
				data:    tt.input,
				readErr: tt.readErr,
			}

			tf := &stubTailFile{
				TailFile: &fileutil.TailFile{
					Reader: bufio.NewReader(reader),
				},
				checkRotationFunc: func() error { return tt.checkRotationErr },
			}

			start := time.Now()
			line, err := tf.NextLine()
			duration := time.Since(start)

			// Allow 200ms tolerance (EOF branch)
			if tt.expectErr == io.EOF && duration < 180*time.Millisecond {
				t.Errorf("NextLine() sleep too short for EOF case")
			}

			if tt.expectErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectErr != nil {
				if !errors.Is(err, tt.expectErr) {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
			}
			if tt.expectLine != "" && string(line) != tt.expectLine {
				t.Errorf("expected line %q, got %q", tt.expectLine, string(line))
			}
		})
	}
}

// faultyReader simulates a controllable reader
type faultyReader struct {
	data    string
	readErr error
	readPos int
}

func (r *faultyReader) Read(p []byte) (int, error) {
	if r.readErr != nil {
		return 0, r.readErr
	}
	if r.readPos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.readPos:])
	r.readPos += n
	return n, nil
}
