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
