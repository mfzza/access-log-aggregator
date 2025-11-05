package logreader

import (
	"fmt"
	"io"
	"os"
)

func seekToLastNLines(f *os.File, n int) {
	if n <= 0 {
		f.Seek(0, io.SeekEnd)
		return
	}

	csize := 512
	stat, err := f.Stat()
	if err != nil {
		return
	}
	fsize := stat.Size()

	var linesFound int
	var offset int64

	for {
		if fsize-int64(csize)-offset < 0 {
			csize = int(fsize - offset)
		}

		offset += int64(csize)
		if offset > fsize {
			offset = fsize
		}

		// Move backward
		pos := max(fsize-offset, 0)
		f.Seek(pos, io.SeekStart)

		tmp := make([]byte, csize)
		f.Read(tmp)

		for i := len(tmp) - 1; i >= 0; i-- {
			if tmp[i] == '\n' {
				linesFound++
				if linesFound > n {
					f.Seek(pos+int64(i+1), io.SeekStart)
					return
				}
			}
		}

		if pos == 0 {
			f.Seek(0, io.SeekEnd)
			return
		}
	}
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

