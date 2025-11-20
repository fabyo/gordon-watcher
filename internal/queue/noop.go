package queue

import (
	"context"

	"github.com/fabyo/gordon-watcher/internal/logger"
)

// NoOpQueue implements Queue interface but does nothing
type NoOpQueue struct {
	logger *logger.Logger
}

// NewNoOpQueue creates a new NoOp queue
func NewNoOpQueue(log *logger.Logger) *NoOpQueue {
	return &NoOpQueue{
		logger: log,
	}
}

// Publish does nothing (no-op)
func (q *NoOpQueue) Publish(ctx context.Context, msg *Message) error {
	q.logger.Debug("NoOp queue: message would be published",
		"messageId", msg.ID,
		"filename", msg.Filename,
		"path", msg.Path,
	)
	return nil
}

// Close does nothing
func (q *NoOpQueue) Close() error {
	q.logger.Info("NoOp queue closed")
	return nil
}