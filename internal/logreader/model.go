package logreader

import (
	"bufio"
	"fmt"
	"os"
)

type Reader struct {
	file *os.File
	buf  *bufio.Reader
}

func NewReader(fp string, fromStart bool) (*Reader, error) {
	file, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %w", err)
	}
	// NOTE: could be added flag -n to more like tail
	if !fromStart{
		seekToLastNLines(file, 10)
	}
	buf := bufio.NewReader(file)
	return &Reader{file: file, buf: buf}, nil
}

func (r *Reader) Close() error {
	return r.file.Close()
}
