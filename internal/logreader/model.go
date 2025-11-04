package logreader

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Reader struct {
	file *os.File
	buf  *bufio.Reader
}

func NewReader(fp string) (*Reader, error) {
	file, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %w", err)
	}
	r := bufio.NewReader(file)
	return &Reader{file: file, buf: r}, nil
}

func (r *Reader) ReadLine() ([]byte,  error) {
	line, err := r.buf.ReadBytes('\n')
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil,  fmt.Errorf("Failed to read line: %w", err)
	}
	return line, nil
}

func (r *Reader) Close() error {
	return r.file.Close()
}
