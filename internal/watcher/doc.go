/*
Package watcher provides a robust file system monitoring and processing pipeline.

Gordon Watcher monitors directories for new files, validates them, ensures idempotency
through SHA256 hashing, and publishes them to a message queue for downstream processing.

# Architecture

The watcher follows an event-driven architecture with the following components:

  - fsnotify integration for real-time file system events
  - Worker pool for concurrent file processing
  - Rate limiter to prevent system overload
  - Stability checker to ensure files are fully written
  - Circuit breaker for resilient queue publishing
  - Distributed locks via Redis for multi-instance coordination

# Basic Usage

	cfg := watcher.Config{
		Paths:             []string{"/data/incoming"},
		FilePatterns:      []string{"*.xml", "*.json"},
		MaxWorkers:        10,
		MaxFilesPerSecond: 100,
		WorkingDir:        "/data",
		Queue:             myQueue,
		Storage:           myStorage,
		Logger:            myLogger,
	}

	w, err := watcher.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := w.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// Graceful shutdown
	defer w.Stop(context.Background())

# File Processing Flow

1. File detected in monitored directory
2. Stability check (waits for file to stop changing)
3. Pattern matching and size validation
4. SHA256 hash calculation
5. Idempotency check (skip if already processed)
6. Distributed lock acquisition
7. Move to processing directory
8. Publish to message queue (with retry + circuit breaker)
9. File remains in processing until external worker completes

# Resilience Features

  - Automatic retry with exponential backoff
  - Circuit breaker to prevent cascading failures
  - Orphan file reconciliation on startup
  - Dead Letter Queue (DLQ) for failed messages
  - Graceful shutdown with in-flight request completion

# Observability

The watcher provides comprehensive observability:

  - Prometheus metrics for monitoring
  - OpenTelemetry tracing for distributed debugging
  - Structured logging with configurable levels
  - Health and readiness endpoints

See the Config type for all available configuration options.
*/
package watcher
