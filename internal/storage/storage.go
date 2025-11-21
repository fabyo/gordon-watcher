package storage

import "context"

// Storage is the interface for state storage
type Storage interface {
	// IsProcessed checks if a file hash has been processed
	IsProcessed(ctx context.Context, hash string) (bool, error)

	// MarkEnqueued marks a file as enqueued for processing
	MarkEnqueued(ctx context.Context, hash, path string) error

	// MarkProcessed marks a file as processed
	MarkProcessed(ctx context.Context, hash string) error

	// MarkFailed marks a file as failed
	MarkFailed(ctx context.Context, hash, reason string) error

	// GetLock acquires a distributed lock for a file
	GetLock(ctx context.Context, hash string) (Lock, error)

	// Close closes the storage connection
	Close() error
}

// Lock represents a distributed lock
type Lock interface {
	// Release releases the lock
	Release(ctx context.Context) error
}
