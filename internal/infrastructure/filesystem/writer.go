package filesystem

import "os"

// FileWriter defines the interface for file system operations
type FileWriter interface {
	WriteFile(path string, content []byte) error
	MkdirAll(path string, perm os.FileMode) error
}

// FileSystemWriter implements FileWriter using os package
type FileSystemWriter struct{}

// NewFileSystemWriter creates a new FileSystemWriter
func NewFileSystemWriter() *FileSystemWriter {
	return &FileSystemWriter{}
}

// WriteFile writes content to a file
func (w *FileSystemWriter) WriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

// MkdirAll creates a directory path
func (w *FileSystemWriter) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
