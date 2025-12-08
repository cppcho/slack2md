package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSystemWriter_WriteFile(t *testing.T) {
	writer := NewFileSystemWriter()

	// Create temp directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Write file
	content := []byte("test content")
	err := writer.WriteFile(filePath, content)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Read back and verify
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Expected content '%s', got '%s'", content, readContent)
	}
}

func TestFileSystemWriter_MkdirAll(t *testing.T) {
	writer := NewFileSystemWriter()

	// Create temp directory
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "a", "b", "c")

	// Create nested directories
	err := writer.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("Directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected path to be a directory")
	}
}

func TestFileSystemWriter_WriteFile_Overwrite(t *testing.T) {
	writer := NewFileSystemWriter()

	// Create temp directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	err := writer.WriteFile(filePath, []byte("initial"))
	if err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	// Overwrite with new content
	newContent := []byte("overwritten")
	err = writer.WriteFile(filePath, newContent)
	if err != nil {
		t.Fatalf("Failed to overwrite file: %v", err)
	}

	// Verify overwritten content
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != string(newContent) {
		t.Errorf("Expected content '%s', got '%s'", newContent, readContent)
	}
}
