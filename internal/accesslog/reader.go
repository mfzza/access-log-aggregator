package accesslog

import (
	"fmt"
	"io"
)

func (lr *Reader) GetRawRecord() ([]byte,  error) {
	line, err := lr.reader.ReadBytes('\n')
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil,  fmt.Errorf("failed to read line: %w", err)
	}
	return line, nil
}

