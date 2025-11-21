package watcher

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool_Submit(t *testing.T) {
	processed := make([]string, 0)
	var mu sync.Mutex

	processFunc := func(ctx context.Context, path string) error {
		mu.Lock()
		processed = append(processed, path)
		mu.Unlock()
		return nil
	}

	pool := NewWorkerPool(2, 10, processFunc)
	pool.Start()
	defer pool.Stop()

	// Submit some files
	pool.Submit("file1.txt")
	pool.Submit("file2.txt")
	pool.Submit("file3.txt")

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := len(processed)
	mu.Unlock()

	if count != 3 {
		t.Errorf("Expected 3 files processed, got %d", count)
	}
}

func TestWorkerPool_SubmitBlocking(t *testing.T) {
	var processCount atomic.Int32

	processFunc := func(ctx context.Context, path string) error {
		processCount.Add(1)
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	// Small queue to test blocking
	pool := NewWorkerPool(1, 2, processFunc)
	pool.Start()
	defer pool.Stop()

	// Submit more files than queue can hold
	for i := 0; i < 5; i++ {
		pool.SubmitBlocking("file.txt")
	}

	// Wait for all to process
	time.Sleep(500 * time.Millisecond)

	if processCount.Load() != 5 {
		t.Errorf("Expected 5 files processed, got %d", processCount.Load())
	}
}

func TestWorkerPool_ConcurrentSubmit(t *testing.T) {
	var processCount atomic.Int32

	processFunc := func(ctx context.Context, path string) error {
		processCount.Add(1)
		return nil
	}

	pool := NewWorkerPool(5, 100, processFunc)
	pool.Start()
	defer pool.Stop()

	// Submit from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				pool.Submit("file.txt")
			}
		}()
	}

	wg.Wait()
	time.Sleep(1 * time.Second)

	if processCount.Load() != 100 {
		t.Errorf("Expected 100 files processed, got %d", processCount.Load())
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	var processCount atomic.Int32

	processFunc := func(ctx context.Context, path string) error {
		processCount.Add(1)
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	pool := NewWorkerPool(2, 10, processFunc)
	pool.Start()

	// Submit files
	for i := 0; i < 5; i++ {
		pool.Submit("file.txt")
	}

	// Stop with timeout
pool.Stop()

	// All submitted files should be processed
	if processCount.Load() != 5 {
		t.Errorf("Expected 5 files processed during shutdown, got %d", processCount.Load())
	}
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	var startedCount atomic.Int32
	var completedCount atomic.Int32

	processFunc := func(ctx context.Context, path string) error {
		startedCount.Add(1)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			completedCount.Add(1)
			return nil
		}
	}

	pool := NewWorkerPool(2, 10, processFunc)
	pool.Start()

	// Submit files
	for i := 0; i < 5; i++ {
		pool.Submit("file.txt")
	}

	// Stop immediately (short timeout)
pool.Stop()

	// Some should have started but not all completed
	if startedCount.Load() == 0 {
		t.Error("Expected some files to start processing")
	}
	if completedCount.Load() >= 5 {
		t.Error("Expected not all files to complete (context cancelled)")
	}
}

func TestWorkerPool_DropWhenFull(t *testing.T) {
	var processCount atomic.Int32

	processFunc := func(ctx context.Context, path string) error {
		processCount.Add(1)
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	// Very small queue
	pool := NewWorkerPool(1, 2, processFunc)
	pool.Start()
	defer pool.Stop()

	// Submit more than queue can hold (non-blocking)
	submitted := 0
	for i := 0; i < 10; i++ {
		pool.Submit("file.txt")
		submitted++
	}

	time.Sleep(2 * time.Second)

	// Not all should be processed (some dropped)
	processed := processCount.Load()
	if processed >= int32(submitted) {
		t.Errorf("Expected some files to be dropped, submitted %d, processed %d", submitted, processed)
	}
}
