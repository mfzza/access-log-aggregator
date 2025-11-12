package tailer

import (
	"fmt"
	"io"
	"os"
)

func (t *TailFile) Close() error {
	return t.file.Close()
}

func (t *TailFile) GetRawRecord() ([]byte, error) {
	line, err := t.reader.ReadBytes('\n')
	if err == io.EOF {
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
	currStat, err := t.fs.Stat(t.fpath)
	if err != nil {
		return nil // dont break the loop
	}

	switch whichRotation(t.fstat, currStat) {
	case truncated:
		t.file.Seek(0, io.SeekStart)
		t.reader.Reset(t.file)
		t.fstat = currStat
	case renamed:
		if t.rotated {
			t.file.Close()
			newFile, err := t.fs.Open(t.fpath)
			if err != nil {
				return fmt.Errorf("reopen file: %w", err)
			}
			t.file = newFile
			t.reader.Reset(newFile)
			t.fstat = currStat

			t.rotated = false
		}
		// first, just mark rotation, let it drain old file
		t.rotated = true
	}
	return nil
}
