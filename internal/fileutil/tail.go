package fileutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

type TailFile struct {
	fpath    string
	file     *os.File
	fileInfo os.FileInfo
	reader   *bufio.Reader
	rotated  bool
}

func NewTailFile(fpath string, fromStart bool) (*TailFile, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	if !fromStart {
		// seekToLastNLines(f, 10)
		f.Seek(0, io.SeekEnd)
	}
	stat, _ := f.Stat()
	return &TailFile{fpath: fpath, file: f, fileInfo: stat, reader: bufio.NewReader(f)}, nil
}

func (t *TailFile) NextLine() ([]byte, error) {
	line, err := t.reader.ReadBytes('\n')
	if err == io.EOF {
		time.Sleep(200 * time.Millisecond)
		if err := t.checkRotation(); err != nil {
			return nil, fmt.Errorf("failed to detect log file rotation: %w", err)
		}
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}
	return line, nil
}
func (t *TailFile) checkRotation() error {
	current, err := os.Stat(t.fpath)
	if err != nil {
		// TODO: log this
		return nil // don’t break the loop
	}

	// if file inode still same
	if os.SameFile(current, t.fileInfo) {
		// and current file size smaller
		if current.Size() < t.fileInfo.Size() {
			// TODO: log this too
			// fmt.Println("Detected truncation, resetting file offset")
			t.file.Seek(0, io.SeekStart)
			t.fileInfo = current
		}
		// if file inode change
	} else {
		// TODO: log this too
		// fmt.Println("Detected rename, reopening file")
		// after old file drained, proceed to new file
		if t.rotated {
			t.file.Close()
			f, err := os.Open(t.fpath)
			if err != nil {
				return fmt.Errorf("cant open new log file")
			}
			t.file = f
			t.fileInfo = current
			t.reader.Reset(f)
			t.rotated = false
		}
		// first, just mark rotation, let it drain old file
		t.rotated = true
	}
	return nil
}
