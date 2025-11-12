// test_doubles.go
package tailer

import (
	"bytes"
	"os"
	"time"
)

// mockFile implements file interface for testing
type mockFile struct {
	*bytes.Reader
	closed    bool
	statFunc  func() (os.FileInfo, error)
	closeFunc func() error
	seekFunc  func(offset int64, whence int) (int64, error)
}

func newMockFile(data []byte) *mockFile {
	return &mockFile{
		Reader: bytes.NewReader(data),
		closed: false,
	}
}

func (m *mockFile) Close() error {
	m.closed = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockFile) Stat() (os.FileInfo, error) {
	if m.statFunc != nil {
		return m.statFunc()
	}
	return &mockFileInfo{name: "test.log", size: int64(m.Len())}, nil
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	if m.seekFunc != nil {
		return m.seekFunc(offset, whence)
	}
	return m.Reader.Seek(offset, whence)
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() any           { return nil }

// mockFileSystem implements fileSystem for testing
type mockFileSystem struct {
	files    map[string]*mockFile
	statFunc func(name string) (os.FileInfo, error)
	openFunc func(name string) (file, error)
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files: make(map[string]*mockFile),
	}
}

func (m *mockFileSystem) Open(name string) (file, error) {
	if m.openFunc != nil {
		return m.openFunc(name)
	}
	if file, exists := m.files[name]; exists {
		return file, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.statFunc != nil {
		return m.statFunc(name)
	}
	if file, exists := m.files[name]; exists {
		return file.Stat()
	}
	return nil, os.ErrNotExist
}
