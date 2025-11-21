package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStabilityChecker_FileBecomesStable(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "stability-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create stability checker
	checker := NewStabilityChecker(3, 100*time.Millisecond)

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait for file to become stable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stable := checker.WaitForStability(ctx, testFile)
	if !stable {
		t.Error("Expected file to become stable")
	}
}

func TestStabilityChecker_FileKeepsChanging(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "stability-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create stability checker with short timeout
	checker := NewStabilityChecker(3, 50*time.Millisecond)

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Keep modifying the file in background
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				os.WriteFile(testFile, []byte(time.Now().String()), 0644)
			}
		}
	}()
	defer close(done)

	// Try to wait for stability (should timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	stable := checker.WaitForStability(ctx, testFile)
	if stable {
		t.Error("Expected file to NOT become stable (keeps changing)")
	}
}

func TestStabilityChecker_ContextCancellation(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "stability-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create stability checker
	checker := NewStabilityChecker(10, 1*time.Second)

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Should return false due to cancelled context
	stable := checker.WaitForStability(ctx, testFile)
	if stable {
		t.Error("Expected false when context is cancelled")
	}
}

func TestStabilityChecker_FileDoesNotExist(t *testing.T) {
	checker := NewStabilityChecker(3, 100*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Try to check stability of non-existent file
	stable := checker.WaitForStability(ctx, "/nonexistent/file.txt")
	if stable {
		t.Error("Expected false for non-existent file")
	}
}

func TestStabilityChecker_RecordDuration(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "stability-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create stability checker
	checker := NewStabilityChecker(2, 100*time.Millisecond)

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	stable := checker.WaitForStability(ctx, testFile)
	duration := time.Since(start)

	if !stable {
		t.Error("Expected file to become stable")
	}

	// Should take at least 1 attempt * 100ms = 100ms
	if duration < 100*time.Millisecond {
		t.Errorf("Expected duration >= 100ms, got %v", duration)
	}
}
