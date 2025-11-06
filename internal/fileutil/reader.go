package fileutil

import (
	"fmt"
	"io"
	"os"
)

func OpenReader(fpath string, fromStart bool) ( *os.File, error ) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %w", err)
	}
	if !fromStart {
		seekToLastNLines(file, 10)
	}
	return file, nil
}

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
