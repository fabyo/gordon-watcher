package integration

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fabyo/gordon-watcher/internal/logger"
	"github.com/fabyo/gordon-watcher/internal/queue"
	"github.com/fabyo/gordon-watcher/internal/storage"
	"github.com/fabyo/gordon-watcher/internal/watcher"
)

// TestEnvironment holds test infrastructure
type TestEnvironment struct {
	TempDir string
	Config  watcher.Config
	Watcher *watcher.Watcher
	Queue   *MockQueue
	Storage *MockStorage
	Logger  *logger.Logger
	Ctx     context.Context
	Cancel  context.CancelFunc
}

// MockQueue implements queue.Queue for testing
type MockQueue struct {
	mu       sync.Mutex
	Messages []*queue.Message
}

func (m *MockQueue) Publish(ctx context.Context, msg *queue.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, msg)
	return nil
}

func (m *MockQueue) Close() error {
	return nil
}

// GetMessageCount returns the number of messages in a thread-safe way
func (m *MockQueue) GetMessageCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Messages)
}

// GetMessage returns a message at index in a thread-safe way
func (m *MockQueue) GetMessage(index int) *queue.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	if index < 0 || index >= len(m.Messages) {
		return nil
	}
	return m.Messages[index]
}

// MockStorage implements storage.Storage for testing
type MockStorage struct {
	mu              sync.Mutex
	ProcessedHashes map[string]bool
	EnqueuedFiles   map[string]string
	FailedFiles     map[string]string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		ProcessedHashes: make(map[string]bool),
		EnqueuedFiles:   make(map[string]string),
		FailedFiles:     make(map[string]string),
	}
}

func (m *MockStorage) IsProcessed(ctx context.Context, hash string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ProcessedHashes[hash], nil
}

func (m *MockStorage) MarkEnqueued(ctx context.Context, hash, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EnqueuedFiles[hash] = path
	return nil
}

func (m *MockStorage) MarkProcessed(ctx context.Context, hash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ProcessedHashes[hash] = true
	return nil
}

func (m *MockStorage) MarkFailed(ctx context.Context, hash, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailedFiles[hash] = reason
	return nil
}

func (m *MockStorage) GetLock(ctx context.Context, hash string) (storage.Lock, error) {
	return &MockLock{}, nil
}

func (m *MockStorage) Close() error {
	return nil
}

// MockLock implements storage.Lock for testing
type MockLock struct{}

func (m *MockLock) Release(ctx context.Context) error {
	return nil
}

// setupTestEnvironment creates a test environment
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "gordon-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create subdirectories
	subdirs := []string{"incoming", "processing", "processed", "failed", "ignored", "tmp"}
	for _, dir := range subdirs {
		if err := os.MkdirAll(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create subdir %s: %v", dir, err)
		}
	}

	// Create logger
	log := logger.New(logger.Config{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	})

	// Create mock queue and storage
	mockQueue := &MockQueue{Messages: make([]*queue.Message, 0)}
	mockStorage := NewMockStorage()

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Create config
	cfg := watcher.Config{
		Paths:             []string{filepath.Join(tempDir, "incoming")},
		FilePatterns:      []string{"*.xml", "*.zip"},
		ExcludePatterns:   []string{".*", "*.tmp"},
		MinFileSize:       10,
		MaxFileSize:       100 * 1024 * 1024,
		StableAttempts:    3,
		StableDelay:       100 * time.Millisecond,
		CleanupInterval:   5 * time.Minute,
		MaxWorkers:        5,
		MaxFilesPerSecond: 100,
		WorkerQueueSize:   10,
		WorkingDir:        tempDir,
		SubDirs: watcher.SubDirectories{
			Processing: "processing",
			Processed:  "processed",
			Failed:     "failed",
			Ignored:    "ignored",
			Tmp:        "tmp",
		},
		Queue:   mockQueue,
		Storage: mockStorage,
		Logger:  log,
	}

	// Create watcher
	w, err := watcher.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	return &TestEnvironment{
		TempDir: tempDir,
		Config:  cfg,
		Watcher: w,
		Queue:   mockQueue,
		Storage: mockStorage,
		Logger:  log,
		Ctx:     ctx,
		Cancel:  cancel,
	}
}

// cleanup removes test environment
func (env *TestEnvironment) cleanup() {
	env.Cancel()
	if env.Watcher != nil {
		_ = env.Watcher.Stop(context.Background())
	}
	_ = os.RemoveAll(env.TempDir)
}

// createTestFile creates a test file with given content
func (env *TestEnvironment) createTestFile(name, content string) (string, error) {
	path := filepath.Join(env.TempDir, "incoming", name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

// createTestXML creates a test XML file
func (env *TestEnvironment) createTestXML(name string) (string, error) {
	content := fmt.Sprintf(`<?xml version="1.0"?><root><item id="%s">Test %s</item></root>`, name, name)
	return env.createTestFile(name, content)
}

// createTestZIP creates a test ZIP file with given files
func (env *TestEnvironment) createTestZIP(zipName string, fileContents map[string]string) (string, error) {
	zipPath := filepath.Join(env.TempDir, "incoming", zipName)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for name, content := range fileContents {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return "", err
		}
		if _, err := io.WriteString(writer, content); err != nil {
			return "", err
		}
	}

	return zipPath, nil
}

// waitForProcessing waits for files to be processed
func (env *TestEnvironment) waitForProcessing(expectedCount int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if env.Queue.GetMessageCount() >= expectedCount {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %d files, got %d", expectedCount, env.Queue.GetMessageCount())
}

// countFilesInDir counts files in a directory
func countFilesInDir(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count, nil
}
