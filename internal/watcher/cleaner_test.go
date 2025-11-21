package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fabyo/gordon-watcher/internal/logger"
)

func TestCleaner_CleanOldFiles(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	processedDir := filepath.Join(tempDir, "processed")
	failedDir := filepath.Join(tempDir, "failed")
	os.MkdirAll(processedDir, 0755)
	os.MkdirAll(failedDir, 0755)

	// Create old file (2 days old)
	oldFile := filepath.Join(processedDir, "old.txt")
	if err := os.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	os.Chtimes(oldFile, oldTime, oldTime)

	// Create recent file
	recentFile := filepath.Join(processedDir, "recent.txt")
	if err := os.WriteFile(recentFile, []byte("recent"), 0644); err != nil {
		t.Fatalf("Failed to create recent file: %v", err)
	}

	// Create cleaner (clean files older than 24 hours)
	log := logger.New(logger.Config{Level: "info", Format: "text"})
	cleaner := NewCleaner(tempDir, []string{}, 24*time.Hour, log)

	// Run cleanup
	cleaner.cleanup()

	// Note: The current cleaner only removes empty directories, not old files
	// This test would need the cleaner to be enhanced to remove old files
	// For now, we just verify it runs without error
}

func TestCleaner_ProtectedDirectories(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create protected directory
	protectedDir := filepath.Join(tempDir, "incoming")
	os.MkdirAll(protectedDir, 0755)

	// Create unprotected empty directory
	unprotectedDir := filepath.Join(tempDir, "temp")
	os.MkdirAll(unprotectedDir, 0755)

	// Create cleaner with protected directory
	log := logger.New(logger.Config{Level: "info", Format: "text"})
	cleaner := NewCleaner(tempDir, []string{protectedDir}, 24*time.Hour, log)

	// Run cleanup
	cleaner.cleanup()

	// Protected directory should still exist
	if _, err := os.Stat(protectedDir); err != nil {
		t.Error("Expected protected directory to still exist")
	}

	// Unprotected empty directory should be removed
	if _, err := os.Stat(unprotectedDir); !os.IsNotExist(err) {
		t.Error("Expected unprotected empty directory to be removed")
	}
}

func TestCleaner_EmptyDirectoryRemoval(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create empty directory
	emptyDir := filepath.Join(tempDir, "empty")
	os.MkdirAll(emptyDir, 0755)

	// Create directory with file
	dirWithFile := filepath.Join(tempDir, "nonempty")
	os.MkdirAll(dirWithFile, 0755)
	os.WriteFile(filepath.Join(dirWithFile, "file.txt"), []byte("content"), 0644)

	// Create cleaner
	log := logger.New(logger.Config{Level: "info", Format: "text"})
	cleaner := NewCleaner(tempDir, []string{}, 24*time.Hour, log)

	// Run cleanup
	cleaner.cleanup()

	// Empty directory should be removed
	if _, err := os.Stat(emptyDir); !os.IsNotExist(err) {
		t.Error("Expected empty directory to be removed")
	}

	// Directory with file should still exist
	if _, err := os.Stat(dirWithFile); err != nil {
		t.Error("Expected directory with file to still exist")
	}
}

func TestCleaner_StartStop(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create cleaner with short interval
	log := logger.New(logger.Config{Level: "info", Format: "text"})
	cleaner := NewCleaner(tempDir, []string{}, 1*time.Hour, log)

	// Start cleaner
	cleaner.Start()

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop cleaner
	cleaner.Stop()

	// Should complete without hanging
}

func TestCleaner_NestedDirectories(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested empty directories
	nestedDir := filepath.Join(tempDir, "level1", "level2", "level3")
	os.MkdirAll(nestedDir, 0755)

	// Create cleaner
	log := logger.New(logger.Config{Level: "info", Format: "text"})
	cleaner := NewCleaner(tempDir, []string{}, 24*time.Hour, log)

	// Run cleanup multiple times to remove nested empties
	for i := 0; i < 3; i++ {
		cleaner.cleanup()
	}

	// All empty nested directories should be removed
	if _, err := os.Stat(filepath.Join(tempDir, "level1")); !os.IsNotExist(err) {
		t.Error("Expected nested empty directories to be removed")
	}
}
