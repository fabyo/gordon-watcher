package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fabyo/gordon-watcher/internal/logger"
	"github.com/fabyo/gordon-watcher/internal/queue"
	"github.com/fabyo/gordon-watcher/internal/storage"
)

// MockQueue implements queue.Queue interface for testing
type MockQueue struct {
	published []*queue.Message
	err       error
}

func (m *MockQueue) Publish(ctx context.Context, msg *queue.Message) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, msg)
	return nil
}

func (m *MockQueue) Close() error {
	return nil
}

// MockStorage implements storage.Storage interface for testing
type MockStorage struct {
	processed map[string]bool
	err       error
}

func (m *MockStorage) IsProcessed(ctx context.Context, hash string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.processed[hash], nil
}

func (m *MockStorage) MarkProcessed(ctx context.Context, hash string) error {
	if m.err != nil {
		return m.err
	}
	m.processed[hash] = true
	return nil
}

func (m *MockStorage) MarkFailed(ctx context.Context, hash string, reason string) error {
	return m.err
}

func (m *MockStorage) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return true, m.err
}

func (m *MockStorage) ReleaseLock(ctx context.Context, key string) error {
	return m.err
}

func (m *MockStorage) Close() error {
	return nil
}

func (m *MockStorage) MarkEnqueued(ctx context.Context, hash, path string) error {
	return m.err
}

type MockLock struct{}

func (ml *MockLock) Release(ctx context.Context) error {
	return nil
}

func (m *MockStorage) GetLock(ctx context.Context, hash string) (storage.Lock, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &MockLock{}, nil
}

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Paths:             []string{tmpDir},
		FilePatterns:      []string{"*.txt"},
		ExcludePatterns:   []string{},
		MinFileSize:       1,
		MaxFileSize:       1024 * 1024,
		StableAttempts:    3,
		StableDelay:       100 * time.Millisecond,
		CleanupInterval:   1 * time.Minute,
		MaxWorkers:        5,
		MaxFilesPerSecond: 10,
		WorkerQueueSize:   10,
		WorkingDir:        tmpDir,
		SubDirs: SubDirectories{
			Processing: "processing",
			Processed:  "processed",
			Failed:     "failed",
			Ignored:    "ignored",
			Tmp:        "tmp",
		},
		Queue:   &MockQueue{},
		Storage: &MockStorage{processed: make(map[string]bool)},
		Logger:  logger.New(logger.Config{Level: "info", Format: "text", Output: "stdout"}),
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	if w == nil {
		t.Fatal("Watcher is nil")
	}

	if w.pool == nil {
		t.Error("Worker pool not initialized")
	}

	if w.rateLimit == nil {
		t.Error("Rate limiter not initialized")
	}

	if w.stability == nil {
		t.Error("Stability checker not initialized")
	}

	if w.cb == nil {
		t.Error("Circuit breaker not initialized")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Paths:             []string{"/tmp"},
				MaxWorkers:        5,
				MaxFilesPerSecond: 10,
				WorkingDir:        "/tmp",
				Queue:             &MockQueue{},
				Storage:           &MockStorage{processed: make(map[string]bool)},
				Logger:            logger.New(logger.Config{Level: "info", Format: "text", Output: "stdout"}),
			},
			wantErr: false,
		},
		{
			name: "no paths",
			cfg: &Config{
				MaxWorkers:        5,
				MaxFilesPerSecond: 10,
				WorkingDir:        "/tmp",
				Queue:             &MockQueue{},
				Storage:           &MockStorage{processed: make(map[string]bool)},
				Logger:            logger.New(logger.Config{Level: "info", Format: "text", Output: "stdout"}),
			},
			wantErr: true,
		},
		{
			name: "invalid max workers",
			cfg: &Config{
				Paths:             []string{"/tmp"},
				MaxWorkers:        0,
				MaxFilesPerSecond: 10,
				WorkingDir:        "/tmp",
				Queue:             &MockQueue{},
				Storage:           &MockStorage{processed: make(map[string]bool)},
				Logger:            logger.New(logger.Config{Level: "info", Format: "text", Output: "stdout"}),
			},
			wantErr: true,
		},
		{
			name: "no queue",
			cfg: &Config{
				Paths:             []string{"/tmp"},
				MaxWorkers:        5,
				MaxFilesPerSecond: 10,
				WorkingDir:        "/tmp",
				Storage:           &MockStorage{processed: make(map[string]bool)},
				Logger:            logger.New(logger.Config{Level: "info", Format: "text", Output: "stdout"}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		pattern  string
		want     bool
	}{
		{"wildcard all", "test.txt", "*", true},
		{"extension match", "test.xml", "*.xml", true},
		{"extension no match", "test.txt", "*.xml", false},
		{"exact match", "test.txt", "test.txt", true},
		{"exact no match", "test.txt", "other.txt", false},
		{"glob pattern", "test123.txt", "test*.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPattern(tt.filename, tt.pattern)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.filename, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestReconcileOrphans(t *testing.T) {
	tmpDir := t.TempDir()
	incomingDir := filepath.Join(tmpDir, "incoming")
	processingDir := filepath.Join(tmpDir, "processing")

	// Create directories
	os.MkdirAll(incomingDir, 0755)
	os.MkdirAll(processingDir, 0755)

	// Create orphan file
	orphanFile := filepath.Join(processingDir, "orphan.txt")
	if err := os.WriteFile(orphanFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create orphan file: %v", err)
	}

	cfg := Config{
		Paths:             []string{incomingDir},
		FilePatterns:      []string{"*.txt"},
		ExcludePatterns:   []string{},
		MinFileSize:       1,
		MaxFileSize:       1024 * 1024,
		StableAttempts:    3,
		StableDelay:       100 * time.Millisecond,
		CleanupInterval:   1 * time.Minute,
		MaxWorkers:        5,
		MaxFilesPerSecond: 10,
		WorkerQueueSize:   10,
		WorkingDir:        tmpDir,
		SubDirs: SubDirectories{
			Processing: "processing",
			Processed:  "processed",
			Failed:     "failed",
			Ignored:    "ignored",
			Tmp:        "tmp",
		},
		Queue:   &MockQueue{},
		Storage: &MockStorage{processed: make(map[string]bool)},
		Logger:  logger.New(logger.Config{Level: "info", Format: "text", Output: "stdout"}),
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	// Run reconciliation
	if err := w.reconcileOrphans(); err != nil {
		t.Fatalf("reconcileOrphans() failed: %v", err)
	}

	// Check orphan was moved
	movedFile := filepath.Join(incomingDir, "orphan.txt")
	if _, err := os.Stat(movedFile); os.IsNotExist(err) {
		t.Error("Orphan file was not moved to incoming")
	}

	// Check processing is empty
	if _, err := os.Stat(orphanFile); !os.IsNotExist(err) {
		t.Error("Orphan file still exists in processing")
	}
}

func TestCalculateHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := Config{
		Paths:             []string{tmpDir},
		FilePatterns:      []string{"*.txt"},
		MaxWorkers:        5,
		MaxFilesPerSecond: 10,
		WorkingDir:        tmpDir,
		SubDirs:           SubDirectories{Processing: "processing"},
		Queue:             &MockQueue{},
		Storage:           &MockStorage{processed: make(map[string]bool)},
		Logger:            logger.New(logger.Config{Level: "info", Format: "text", Output: "stdout"}),
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	hash, err := w.calculateHash(testFile)
	if err != nil {
		t.Fatalf("calculateHash() failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash is empty")
	}

	if len(hash) != 64 { // SHA256 produces 64 hex characters
		t.Errorf("Hash length = %d, want 64", len(hash))
	}

	// Calculate again, should be same
	hash2, err := w.calculateHash(testFile)
	if err != nil {
		t.Fatalf("calculateHash() second call failed: %v", err)
	}

	if hash != hash2 {
		t.Error("Hash is not deterministic")
	}
}
