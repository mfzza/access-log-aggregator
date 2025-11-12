package tailer

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type fileSystem interface {
	Open(name string) (file, error)
	Stat(name string) (os.FileInfo, error)
}

type file interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	Stat() (os.FileInfo, error)
}

// OsFS implements fileSystem using the local disk.
type OsFS struct{}

func (OsFS) Open(name string) (file, error)        { return os.Open(name) }
func (OsFS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

type Tailer interface {
	GetRawRecord() ([]byte, error)
	Close() error
}

type TailFile struct {
	fpath   string
	file    file
	reader  *bufio.Reader
	fstat   os.FileInfo
	rotated bool
	fs      fileSystem
}

func NewTailFile(fpath string, fs fileSystem, fromStart bool) (*TailFile, error) {
	file, err := fs.Open(fpath)
	if err != nil {
		return nil, err
	}

	if !fromStart {
		file.Seek(0, io.SeekEnd)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("get file stat: %w", err)
	}
	return &TailFile{fpath: fpath, file: file, reader: bufio.NewReader(file), fstat: stat, fs: fs}, nil
}

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
