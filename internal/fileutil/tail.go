package fileutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type TailFile struct {
	path      string
	fromStart bool
	f         *os.File
	stat      os.FileInfo
	r         *bufio.Reader
}

func NewTailFile(path string, fromStart bool) (*TailFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	if !fromStart {
		seekToLastNLines(f, 10)
	}
	stat, _ := f.Stat()
	return &TailFile{path: path, fromStart: fromStart, f: f, stat: stat, r: bufio.NewReader(f)}, nil
}

func (t *TailFile) NextLine() ([]byte, error) {
	line, err := t.r.ReadBytes('\n')
	if err == io.EOF {
		if err := t.checkRotation(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}
	return line, nil
}
func (t *TailFile) checkRotation() error {
	current, err := os.Stat(t.path)
	if err != nil {
		return nil // don’t break the loop
	}

	// Truncated
	if current.Size() < t.stat.Size() {
		fmt.Println("Detected truncation, resetting file offset")
		t.f.Seek(0, io.SeekStart)
		t.stat = current
		return nil
	}

	// Rotated
	if !os.SameFile(current, t.stat) {
		fmt.Println("Detected rotation, reopening file")
		t.f.Close()
		f, err := os.Open(t.path)
		if err != nil {
			return err
		}
		t.f = f
		t.stat = current
	}
	return nil
}
