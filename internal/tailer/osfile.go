package tailer

import (
	"io"
	"os"
)

type OSFile struct {
	*os.File
}

func NewOSFile(fpath string, fromStart bool) (*OSFile, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	if !fromStart {
		file.Seek(0, io.SeekEnd)
	}

	return &OSFile{file}, nil
}

func (f *OSFile) Stat() (os.FileInfo, error) {
	return f.File.Stat()
}

func (f *OSFile) NewStat(fpath string) (os.FileInfo, error) {
	return os.Stat(fpath)
}

func (f *OSFile) Open(fpath string) (FileSrc, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	return &OSFile{file}, nil
}
