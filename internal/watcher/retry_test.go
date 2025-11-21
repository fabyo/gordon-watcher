package watcher

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry_Success(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		return nil
	}

	cfg := DefaultRetryConfig()
	ctx := context.Background()

	err := Retry(ctx, cfg, fn)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetry_EventualSuccess(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	cfg := DefaultRetryConfig()
	cfg.InitialDelay = 10 * time.Millisecond
	ctx := context.Background()

	err := Retry(ctx, cfg, fn)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_MaxAttemptsExceeded(t *testing.T) {
	attempts := 0
	testErr := errors.New("persistent error")
	fn := func() error {
		attempts++
		return testErr
	}

	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3
	cfg.InitialDelay = 10 * time.Millisecond
	ctx := context.Background()

	err := Retry(ctx, cfg, fn)
	if err == nil {
		t.Error("Expected error after max attempts")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	attempts := 0
	delays := make([]time.Duration, 0)
	lastTime := time.Now()

	fn := func() error {
		now := time.Now()
		if attempts > 0 {
			delays = append(delays, now.Sub(lastTime))
		}
		lastTime = now
		attempts++
		if attempts < 4 {
			return errors.New("error")
		}
		return nil
	}

	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.InitialDelay = 50 * time.Millisecond
	cfg.MaxDelay = 500 * time.Millisecond
	cfg.Multiplier = 2.0
	ctx := context.Background()

	err := Retry(ctx, cfg, fn)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify exponential backoff
	if len(delays) < 2 {
		t.Fatal("Not enough delays recorded")
	}

	// Second delay should be roughly 2x the first
	if delays[1] < delays[0]*15/10 { // Allow 50% tolerance
		t.Errorf("Expected exponential backoff, got delays: %v", delays)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		time.Sleep(100 * time.Millisecond)
		return errors.New("error")
	}

	cfg := DefaultRetryConfig()
	cfg.InitialDelay = 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := Retry(ctx, cfg, fn)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	// Should have attempted at least once but not completed all retries
	if attempts == 0 {
		t.Error("Expected at least one attempt")
	}
	if attempts >= cfg.MaxAttempts {
		t.Errorf("Expected fewer than %d attempts due to context cancellation, got %d", cfg.MaxAttempts, attempts)
	}
}

func TestRetry_MaxDelayRespected(t *testing.T) {
	attempts := 0
	delays := make([]time.Duration, 0)
	lastTime := time.Now()

	fn := func() error {
		now := time.Now()
		if attempts > 0 {
			delays = append(delays, now.Sub(lastTime))
		}
		lastTime = now
		attempts++
		if attempts < 5 {
			return errors.New("error")
		}
		return nil
	}

	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 6
	cfg.InitialDelay = 50 * time.Millisecond
	cfg.MaxDelay = 100 * time.Millisecond
	cfg.Multiplier = 3.0 // High multiplier
	ctx := context.Background()

	err := Retry(ctx, cfg, fn)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify max delay is respected
	for i, delay := range delays {
		if delay > cfg.MaxDelay*12/10 { // Allow 20% tolerance
			t.Errorf("Delay %d exceeded max delay: %v > %v", i, delay, cfg.MaxDelay)
		}
	}
}

func TestRetry_ZeroAttempts(t *testing.T) {
	fn := func() error {
		return errors.New("error")
	}

	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 0
	ctx := context.Background()

	err := Retry(ctx, cfg, fn)
	if err == nil {
		t.Error("Expected error with zero max attempts")
	}
}
