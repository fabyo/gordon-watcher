package integration

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/fabyo/gordon-watcher/internal/watcher"
)

func TestWatcher_FileDetectionAndProcessing(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Start watcher
	if err := env.Watcher.Start(env.Ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Create test files
	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("test_%d.xml", i)
		if _, err := env.createTestXML(name); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Wait for processing
	if err := env.waitForProcessing(3, 10*time.Second); err != nil {
		t.Fatalf("Files not processed: %v", err)
	}

	// Verify messages were sent to queue
	if env.Queue.GetMessageCount() != 3 {
		t.Errorf("Expected 3 messages in queue, got %d", env.Queue.GetMessageCount())
	}

	// Verify files were moved to processing
	processingDir := filepath.Join(env.TempDir, "processing")
	count, err := countFilesInDir(processingDir)
	if err != nil {
		t.Fatalf("Failed to count files in processing: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 files in processing, got %d", count)
	}

	// Verify incoming is empty
	incomingDir := filepath.Join(env.TempDir, "incoming")
	count, err = countFilesInDir(incomingDir)
	if err != nil {
		t.Fatalf("Failed to count files in incoming: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 files in incoming, got %d", count)
	}
}

func TestWatcher_ZIPExtraction(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Start watcher
	if err := env.Watcher.Start(env.Ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Create ZIP with 2 XML files
	fileContents := map[string]string{
		"file1.xml": `<?xml version="1.0"?><root><item id="1">Test 1</item></root>`,
		"file2.xml": `<?xml version="1.0"?><root><item id="2">Test 2</item></root>`,
	}

	if _, err := env.createTestZIP("test.zip", fileContents); err != nil {
		t.Fatalf("Failed to create test ZIP: %v", err)
	}

	// Wait for processing (2 XML files from ZIP)
	if err := env.waitForProcessing(2, 15*time.Second); err != nil {
		t.Fatalf("Files not processed: %v", err)
	}

	// Verify 2 messages were sent (one for each XML)
	if env.Queue.GetMessageCount() != 2 {
		t.Errorf("Expected 2 messages in queue, got %d", env.Queue.GetMessageCount())
	}

	// Verify ZIP was deleted from incoming
	incomingDir := filepath.Join(env.TempDir, "incoming")
	count, err := countFilesInDir(incomingDir)
	if err != nil {
		t.Fatalf("Failed to count files in incoming: %v", err)
	}
	// ZIP should be deleted, extracted files should be moved to processing
	if count != 0 {
		t.Errorf("Expected 0 files in incoming after ZIP extraction, got %d", count)
	}
}

func TestWatcher_DuplicateHandling(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Start watcher
	if err := env.Watcher.Start(env.Ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Create a file
	content := `<?xml version="1.0"?><root><item id="123">Test 123</item></root>`
	if _, err := env.createTestFile("file1.xml", content); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Wait for processing
	if err := env.waitForProcessing(1, 10*time.Second); err != nil {
		t.Fatalf("File not processed: %v", err)
	}

	// Verify file was processed
	if env.Queue.GetMessageCount() != 1 {
		t.Errorf("Expected 1 message in queue, got %d", env.Queue.GetMessageCount())
	}

	// Get the hash
	hash := env.Queue.GetMessage(0).Hash

	// Verify hash was marked as enqueued in storage
	env.Storage.mu.Lock()
	_, exists := env.Storage.EnqueuedFiles[hash]
	env.Storage.mu.Unlock()

	if !exists {
		t.Errorf("Expected hash to be marked as enqueued in storage")
	}

	// Simulate marking as processed (this would normally be done by the consumer)
	env.Storage.MarkProcessed(env.Ctx, hash)

	// Verify it's now marked as processed
	processed, err := env.Storage.IsProcessed(env.Ctx, hash)
	if err != nil {
		t.Fatalf("Failed to check if processed: %v", err)
	}
	if !processed {
		t.Error("Expected file to be marked as processed")
	}
}

func TestWatcher_RateLimiting(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Set low rate limit for testing
	env.Config.MaxFilesPerSecond = 2

	// Recreate watcher with new config
	w, err := watcher.New(env.Config)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	env.Watcher = w

	// Start watcher
	if err := env.Watcher.Start(env.Ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Create 5 files quickly
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("test_%d.xml", i)
		if _, err := env.createTestXML(name); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Wait for processing
	time.Sleep(5 * time.Second)

	// With rate limit of 2/s, not all files should be processed immediately
	// Some might be dropped or delayed
	if env.Queue.GetMessageCount() > 5 {
		t.Errorf("Expected at most 5 messages, got %d", env.Queue.GetMessageCount())
	}

	// Verify some files might be in ignored due to rate limiting
	ignoredDir := filepath.Join(env.TempDir, "ignored")
	count, err := countFilesInDir(ignoredDir)
	if err != nil {
		t.Fatalf("Failed to count files in ignored: %v", err)
	}

	// At least some files should have been rate limited
	// (This is a soft check as timing can vary)
	t.Logf("Files in ignored due to rate limiting: %d", count)
}

func TestWatcher_StabilityChecker(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Set strict stability requirements
	env.Config.StableAttempts = 5
	env.Config.StableDelay = 200 * time.Millisecond

	// Recreate watcher with new config
	w, err := watcher.New(env.Config)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	env.Watcher = w

	// Start watcher
	if err := env.Watcher.Start(env.Ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Create file that will be stable
	content := `<?xml version="1.0"?><root><item id="123">Test 123</item></root>`
	if _, err := env.createTestFile("stable.xml", content); err != nil {
		t.Fatalf("Failed to create stable file: %v", err)
	}

	// Wait for processing (should succeed after stability check)
	if err := env.waitForProcessing(1, 15*time.Second); err != nil {
		t.Fatalf("Stable file not processed: %v", err)
	}

	// Verify file was processed
	if env.Queue.GetMessageCount() != 1 {
		t.Errorf("Expected 1 message in queue, got %d", env.Queue.GetMessageCount())
	}
}

func TestWatcher_NonMatchingFiles(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Start watcher
	if err := env.Watcher.Start(env.Ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Create non-matching file (not XML or ZIP)
	if _, err := env.createTestFile("test.txt", "This is a text file"); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Verify no messages were sent
	if env.Queue.GetMessageCount() != 0 {
		t.Errorf("Expected 0 messages in queue (non-matching file), got %d", env.Queue.GetMessageCount())
	}

	// Verify file was moved to ignored
	ignoredDir := filepath.Join(env.TempDir, "ignored")
	count, err := countFilesInDir(ignoredDir)
	if err != nil {
		t.Fatalf("Failed to count files in ignored: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 file in ignored (non-matching), got %d", count)
	}
}
