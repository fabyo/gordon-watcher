package watcher

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimiter limits the rate of file processing
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxFilesPerSecond int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(maxFilesPerSecond), maxFilesPerSecond),
	}
}

// Allow checks if an operation is allowed under the rate limit
func (r *RateLimiter) Allow() bool {
	return r.limiter.Allow()
}

// Wait waits until the rate limit allows another operation
func (r *RateLimiter) Wait() {
	_ = r.limiter.Wait(context.Background())
}
