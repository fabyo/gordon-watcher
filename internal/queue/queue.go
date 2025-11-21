package queue

import (
	"context"
	"time"
)

// Queue is the interface for message queue
type Queue interface {
	// Publish publishes a message to the queue
	Publish(ctx context.Context, msg *Message) error

	// Close closes the queue connection
	Close() error
}

// Message represents a file event message
type Message struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Filename  string    `json:"filename"`
	Kind      string    `json:"kind"`
	Size      int64     `json:"size"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
}
