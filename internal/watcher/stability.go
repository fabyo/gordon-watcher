package watcher

import (
	"context"
	"os"
	"time"

	"github.com/fabyo/gordon-watcher/internal/metrics"
)

// StabilityChecker checks if a file has stabilized (stopped changing)
type StabilityChecker struct {
	attempts int
	delay    time.Duration
}

// NewStabilityChecker creates a new stability checker
func NewStabilityChecker(attempts int, delay time.Duration) *StabilityChecker {
	return &StabilityChecker{
		attempts: attempts,
		delay:    delay,
	}
}

// WaitForStability waits for a file to stabilize
func (s *StabilityChecker) WaitForStability(ctx context.Context, path string) bool {
	startTime := time.Now()

	var lastSize int64
	var lastModTime time.Time

	for i := 0; i < s.attempts; i++ {
		// Get file info
		info, err := os.Stat(path)
		if err != nil {
			// File was deleted or inaccessible
			return false
		}

		// First iteration
		if i == 0 {
			lastSize = info.Size()
			lastModTime = info.ModTime()

			select {
			case <-ctx.Done():
				return false
			case <-time.After(s.delay):
				continue
			}
		}

		// Check if file changed
		if info.Size() == lastSize && info.ModTime().Equal(lastModTime) {
			// File is stable!
			duration := time.Since(startTime)
			metrics.FileStabilityDuration.Observe(duration.Seconds())
			return true
		}

		// Update and wait
		lastSize = info.Size()
		lastModTime = info.ModTime()

		select {
		case <-ctx.Done():
			return false
		case <-time.After(s.delay):
			// Continue
		}
	}

	// File did not stabilize
	return false
}
