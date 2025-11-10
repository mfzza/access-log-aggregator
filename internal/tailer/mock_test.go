package tailer

import (
	"io"
	"os"
	"time"
)

// MockFileInfo for testing
type MockFileInfo struct {
	NameVal    string
	SizeVal    int64
	ModTimeVal time.Time
	ModeVal    os.FileMode
	IsDirVal   bool
}

func (m *MockFileInfo) Name() string       { return m.NameVal }
func (m *MockFileInfo) Size() int64        { return m.SizeVal }
func (m *MockFileInfo) Mode() os.FileMode  { return m.ModeVal }
func (m *MockFileInfo) ModTime() time.Time { return m.ModTimeVal }
func (m *MockFileInfo) IsDir() bool        { return m.IsDirVal }
func (m *MockFileInfo) Sys() any   { return nil }

// MockFile for testing
type MockFile struct {
	Content      []byte
	ReadPos      int
	Closed       bool
	SeekPos      int64
	FilePath     string
	StatFunc     func() (os.FileInfo, error)
	NewStatFunc  func(string) (os.FileInfo, error)
	OpenFunc     func(string) (FileSrc, error)
	SeekFunc     func(int64, int) (int64, error)
}

func NewMockFile(content []byte, filePath string) *MockFile {
	return &MockFile{
		Content:  content,
		FilePath: filePath,
		StatFunc: func() (os.FileInfo, error) {
			return &MockFileInfo{
				NameVal:    filePath,
				SizeVal:    int64(len(content)),
				ModTimeVal: time.Now(),
				ModeVal:    0644,
			}, nil
		},
		NewStatFunc: func(path string) (os.FileInfo, error) {
			return &MockFileInfo{
				NameVal:    path,
				SizeVal:    int64(len(content)),
				ModTimeVal: time.Now(),
				ModeVal:    0644,
			}, nil
		},
		OpenFunc: func(path string) (FileSrc, error) {
			return NewMockFile([]byte("new file content"), path), nil
		},
	}
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	if m.Closed {
		return 0, os.ErrClosed
	}
	if m.ReadPos >= len(m.Content) {
		return 0, io.EOF
	}
	n = copy(p, m.Content[m.ReadPos:])
	m.ReadPos += n
	return n, nil
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	if m.SeekFunc != nil {
		return m.SeekFunc(offset, whence)
	}

	switch whence {
	case io.SeekStart:
		m.SeekPos = offset
	case io.SeekCurrent:
		m.SeekPos += offset
	case io.SeekEnd:
		m.SeekPos = int64(len(m.Content)) + offset
	}
	m.ReadPos = min(max(int(m.SeekPos), 0), len(m.Content))
	return m.SeekPos, nil
}

func (m *MockFile) Close() error {
	m.Closed = true
	return nil
}

func (m *MockFile) Stat() (os.FileInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc()
	}
	return &MockFileInfo{
		NameVal:    m.FilePath,
		SizeVal:    int64(len(m.Content)),
		ModTimeVal: time.Now(),
	}, nil
}

func (m *MockFile) NewStat(fpath string) (os.FileInfo, error) {
	if m.NewStatFunc != nil {
		return m.NewStatFunc(fpath)
	}
	return &MockFileInfo{
		NameVal:    fpath,
		SizeVal:    int64(len(m.Content)),
		ModTimeVal: time.Now(),
	}, nil
}

func (m *MockFile) Open(name string) (FileSrc, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(name)
	}
	return NewMockFile([]byte("new file content"), name), nil
}
