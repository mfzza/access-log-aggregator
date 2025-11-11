package tailer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

type FileSrc interface {
	io.Reader
	io.Seeker
	io.Closer
	Stat() (os.FileInfo, error)
	NewStat(fpath string) (os.FileInfo, error)
	Open(name string) (FileSrc, error)
}

type TailFile struct {
	fpath   string
	file    FileSrc
	reader  *bufio.Reader
	fstat   os.FileInfo
	rotated bool
	delay   time.Duration
}

func NewTailFile(fpath string, file FileSrc, delay time.Duration) (*TailFile, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("get file stat: %w", err)
	}
	return &TailFile{fpath: fpath, file: file, reader: bufio.NewReader(file), fstat: stat, delay: delay}, nil
}

func (t *TailFile) Close() error {
	return t.file.Close()
}

func (t *TailFile) GetRawRecord() ([]byte, error) {
	line, err := t.reader.ReadBytes('\n')
	if err == io.EOF {
		time.Sleep(t.delay)
		if err := t.checkRotation(); err != nil {
			return nil, fmt.Errorf("detect rotation: %w", err)
		}
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("read line: %w", err)
	}
	return line, nil
}

type rotationStatus int

const (
	same rotationStatus = iota
	truncated
	renamed
)

var sameFile = os.SameFile

func whichRotation(lastStat, currStat os.FileInfo) rotationStatus {
	if sameFile(lastStat, currStat) {
		if currStat.Size() < lastStat.Size() {
			return truncated
		}
		return same
	}
	return renamed
}

func (t *TailFile) checkRotation() error {
	currStat, err := t.file.NewStat(t.fpath)
	if err != nil {
		return nil // dont break the loop
	}

	switch whichRotation(t.fstat, currStat) {
	case truncated:
		t.file.Seek(0, io.SeekStart)
		t.reader.Reset(t.file)
	case renamed:
		if t.rotated {
			t.file.Close()
			newFile, err := t.file.Open(t.fpath)
			if err != nil {
				return fmt.Errorf("reopen file: %w", err)
			}
			t.file = newFile
			t.reader.Reset(newFile)

			t.rotated = false
		}
		// first, just mark rotation, let it drain old file
		t.rotated = true
	}
	t.fstat = currStat
	return nil
}
