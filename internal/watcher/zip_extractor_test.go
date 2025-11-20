package watcher

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZip(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "zip_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a ZIP file
	zipPath := filepath.Join(tmpDir, "test.zip")
	createTestZip(t, zipPath)

	// Create destination directory
	destDir := filepath.Join(tmpDir, "extracted")

	// Test extraction
	extractedFiles, err := ExtractZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("ExtractZip failed: %v", err)
	}

	// Verify extracted files
	expectedFiles := []string{
		filepath.Join(destDir, "file1.txt"),
		filepath.Join(destDir, "subdir", "file2.txt"),
	}

	if len(extractedFiles) != len(expectedFiles) {
		t.Errorf("Expected %d extracted files, got %d", len(expectedFiles), len(extractedFiles))
	}

	for _, expected := range expectedFiles {
		if _, err := os.Stat(expected); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", expected)
		}
	}

	// Verify content
	content1, _ := os.ReadFile(filepath.Join(destDir, "file1.txt"))
	if string(content1) != "content1" {
		t.Errorf("Expected file1 content to be 'content1', got '%s'", string(content1))
	}
}

func TestIsZipFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"test.zip", true},
		{"TEST.ZIP", true},
		{"test.xml", false},
		{"test.zip.tmp", false},
	}

	for _, tt := range tests {
		if got := IsZipFile(tt.filename); got != tt.want {
			t.Errorf("IsZipFile(%s) = %v, want %v", tt.filename, got, tt.want)
		}
	}
}

func createTestZip(t *testing.T, path string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	// Add file 1
	f1, err := w.Create("file1.txt")
	if err != nil {
		t.Fatalf("Failed to create file1 in zip: %v", err)
	}
	f1.Write([]byte("content1"))

	// Add file 2 in subdir
	f2, err := w.Create("subdir/file2.txt")
	if err != nil {
		t.Fatalf("Failed to create file2 in zip: %v", err)
	}
	f2.Write([]byte("content2"))
}
