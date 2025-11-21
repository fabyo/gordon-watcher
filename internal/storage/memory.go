package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryStorage implements Storage using in-memory maps
type MemoryStorage struct {
	mu sync.RWMutex

	processed map[string]time.Time
	enqueued  map[string]string
	failed    map[string]string
	locks     map[string]*memoryLock
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		processed: make(map[string]time.Time),
		enqueued:  make(map[string]string),
		failed:    make(map[string]string),
		locks:     make(map[string]*memoryLock),
	}
}

// IsProcessed checks if a file hash has been processed
func (s *MemoryStorage) IsProcessed(ctx context.Context, hash string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.processed[hash]
	return exists, nil
}

// MarkEnqueued marks a file as enqueued
func (s *MemoryStorage) MarkEnqueued(ctx context.Context, hash, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enqueued[hash] = path
	return nil
}

// MarkProcessed marks a file as processed
func (s *MemoryStorage) MarkProcessed(ctx context.Context, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processed[hash] = time.Now()
	delete(s.enqueued, hash)

	return nil
}

// MarkFailed marks a file as failed
func (s *MemoryStorage) MarkFailed(ctx context.Context, hash, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failed[hash] = reason
	delete(s.enqueued, hash)

	return nil
}

// GetLock acquires a lock
func (s *MemoryStorage) GetLock(ctx context.Context, hash string) (Lock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.locks[hash]; exists {
		return nil, fmt.Errorf("lock already held")
	}

	lock := &memoryLock{
		storage: s,
		hash:    hash,
	}

	s.locks[hash] = lock
	return lock, nil
}

// Close does nothing for memory storage
func (s *MemoryStorage) Close() error {
	return nil
}

// memoryLock implements Lock for in-memory storage
type memoryLock struct {
	storage *MemoryStorage
	hash    string
}

// Release releases the lock
func (l *memoryLock) Release(ctx context.Context) error {
	l.storage.mu.Lock()
	defer l.storage.mu.Unlock()

	delete(l.storage.locks, l.hash)
	return nil
}
