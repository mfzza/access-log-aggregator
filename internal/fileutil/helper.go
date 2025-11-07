package fileutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// FIXME: should return max line if file lines is < n
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

func NextLine(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadBytes('\n')
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}
	return line, nil
}

type LogRotate struct {
	stat      os.FileInfo
	Truncated bool
	Renamed   bool
}

func NewLogRotate(fpath string) (*LogRotate, error) {
	fstat, err := os.Stat(fpath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stat: %w", err)
	}
	return &LogRotate{stat: fstat, Truncated: false, Renamed: false}, nil
}

func (lr *LogRotate) CheckRotation(fpath string) error {
	currStat, err := os.Stat(fpath)
	if err != nil {
		return nil // don’t break the loop
	}

	// if file inode still same
	if os.SameFile(currStat, lr.stat) {
		// and current file size smaller
		if currStat.Size() < lr.stat.Size() {
			lr.stat = currStat
			lr.Renamed = false
			lr.Truncated = true
			return nil
			// NOTE: return FileInfo, bool seek
		}
	} else {
		lr.stat = currStat
		lr.Renamed = true
		lr.Truncated = false
		return nil
	}
	return nil
}
