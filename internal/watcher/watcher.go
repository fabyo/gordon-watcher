package watcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/fabyo/gordon-watcher/internal/logger"
	"github.com/fabyo/gordon-watcher/internal/metrics"
	"github.com/fabyo/gordon-watcher/internal/queue"
	"github.com/fabyo/gordon-watcher/internal/storage"
)

// Config holds watcher configuration
type Config struct {
	// Paths to watch
	Paths []string

	// File patterns to match
	FilePatterns []string

	// Patterns to exclude
	ExcludePatterns []string

	// File size constraints (bytes)
	MinFileSize int64
	MaxFileSize int64

	// Stability check settings
	StableAttempts int
	StableDelay    time.Duration

	// Cleanup interval for empty directories
	CleanupInterval time.Duration

	// Worker pool settings
	MaxWorkers        int
	MaxFilesPerSecond int
	WorkerQueueSize   int

	// Working directory
	WorkingDir string

	// Subdirectories
	SubDirs SubDirectories

	// Dependencies
	Queue   queue.Queue
	Storage storage.Storage
	Logger  *logger.Logger
}

// SubDirectories defines directory structure
type SubDirectories struct {
	Processing string
	Processed  string
	Failed     string
	Ignored    string
	Tmp        string
}

// Watcher monitors file system events
type Watcher struct {
	cfg       Config
	fsWatcher *fsnotify.Watcher
	pool      *WorkerPool
	rateLimit *RateLimiter
	cleaner   *Cleaner
	stability *StabilityChecker
	cb        *CircuitBreaker

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	tracer trace.Tracer

	// Track ignored files to deduplicate metric
	ignoredFiles sync.Map // map[string]time.Time
}

// New creates a new Watcher instance
func New(cfg Config) (*Watcher, error) {
	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create fsnotify watcher
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	w := &Watcher{
		cfg:       cfg,
		fsWatcher: fsWatcher,
		ctx:       ctx,
		cancel:    cancel,
		tracer:    otel.Tracer("gordon-watcher"),
	}

	// Initialize components
	w.pool = NewWorkerPool(cfg.MaxWorkers, cfg.WorkerQueueSize, w.processFile)
	w.rateLimit = NewRateLimiter(cfg.MaxFilesPerSecond)
	w.cleaner = NewCleaner(cfg.WorkingDir, cfg.CleanupInterval, cfg.Logger)
	w.stability = NewStabilityChecker(cfg.StableAttempts, cfg.StableDelay)
	w.cb = NewCircuitBreaker(5, 30*time.Second) // 5 failures, 30s reset timeout

	return w, nil
}

// Start starts the watcher
func (w *Watcher) Start(ctx context.Context) error {
	w.cfg.Logger.Info("Starting Gordon Watcher",
		"paths", w.cfg.Paths,
		"workers", w.cfg.MaxWorkers,
		"rateLimit", w.cfg.MaxFilesPerSecond,
	)

	// Create necessary directories
	if err := w.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Reconcile orphan files in processing directory
	if err := w.reconcileOrphans(); err != nil {
		w.cfg.Logger.Error("Failed to reconcile orphans", "error", err)
		// Continue anyway, don't block startup
	}

	// Start worker pool
	w.pool.Start()

	// Start cleaner
	w.cleaner.Start()

	// Start directory monitor to ensure directories don't disappear
	w.wg.Add(1)
	go w.monitorDirectories()

	// Add watch paths
	for _, path := range w.cfg.Paths {
		w.cfg.Logger.Info("Adding watch path", "path", path)

		// Scan existing files
		if err := w.scanAndWatch(path); err != nil {
			w.cfg.Logger.Error("Failed to scan path", "path", path, "error", err)
		}
	}

	// Start event loop
	w.wg.Add(1)
	go w.eventLoop()

	w.cfg.Logger.Info("Gordon Watcher started successfully")

	return nil
}

// Stop stops the watcher gracefully
func (w *Watcher) Stop(ctx context.Context) error {
	w.cfg.Logger.Info("Stopping Gordon Watcher...")

	// Cancel context
	w.cancel()

	// Close fsnotify watcher
	if err := w.fsWatcher.Close(); err != nil {
		w.cfg.Logger.Error("Error closing fsnotify watcher", "error", err)
	}

	// Stop worker pool
	w.pool.Stop()

	// Stop cleaner
	w.cleaner.Stop()

	// Wait for event loop to finish
	w.wg.Wait()

	w.cfg.Logger.Info("Gordon Watcher stopped")

	return nil
}

// eventLoop processes file system events
func (w *Watcher) eventLoop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			w.handleEvent(event)

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}

			w.cfg.Logger.Error("Watcher error", "error", err)
			metrics.WatcherErrors.Inc()
		}
	}
}

// handleEvent handles a file system event
func (w *Watcher) handleEvent(event fsnotify.Event) {
	ctx, span := w.tracer.Start(w.ctx, "handleEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("event.path", event.Name),
		attribute.String("event.op", event.Op.String()),
	)

	// Only process Create and Write events
	if event.Op&fsnotify.Create != fsnotify.Create &&
		event.Op&fsnotify.Write != fsnotify.Write {
		return
	}

	// Check if it's a directory
	info, err := os.Stat(event.Name)
	if err != nil {
		// File might have been deleted already
		return
	}

	if info.IsDir() {
		// Add directory to watcher
		if err := w.fsWatcher.Add(event.Name); err != nil {
			w.cfg.Logger.Error("Failed to add directory to watcher",
				"path", event.Name,
				"error", err)
		} else {
			w.cfg.Logger.Debug("Directory added to watcher", "path", event.Name)
		}
		return
	}

	// Auto-delete Windows Zone.Identifier files
	if strings.HasSuffix(event.Name, ":Zone.Identifier") || strings.HasSuffix(event.Name, ".Zone.Identifier") {
		w.cfg.Logger.Debug("Deleting Zone.Identifier file", "path", event.Name)
		if err := os.Remove(event.Name); err != nil {
			w.cfg.Logger.Warn("Failed to delete Zone.Identifier", "path", event.Name, "error", err)
		}
		return
	}

	// Check if file matches patterns
	if !w.matchesPatterns(event.Name) {
		w.cfg.Logger.Debug("File does not match patterns", "path", event.Name)
		// Deduplicate ignored files metric (only count once per file)
		if _, exists := w.ignoredFiles.LoadOrStore(event.Name, time.Now()); !exists {
			metrics.FilesIgnored.Inc()
		}
		return
	}

	w.cfg.Logger.Info("File detected", "path", event.Name)

	// Launch goroutine for stability check and processing
	// This prevents blocking the event loop while waiting for file stability
	w.wg.Add(1)
	go func(path string, parentCtx context.Context) {
		defer w.wg.Done()

		// Create a new context for this operation that respects the trace context
		// but can also be cancelled independently if needed
		ctx, cancel := context.WithTimeout(parentCtx, 5*time.Minute) // Safety timeout
		defer cancel()

		// Wait for file to stabilize
		if !w.stability.WaitForStability(ctx, path) {
			w.cfg.Logger.Warn("File did not stabilize", "path", path)
			w.moveToIgnored(path, "file_not_stable")
			return
		}

		// Increment metric only after file is stable and ready for processing
		metrics.FilesDetected.Inc()

		// Apply rate limiting
		if !w.rateLimit.Allow() {
			w.cfg.Logger.Warn("Rate limit exceeded, dropping file", "path", path)
			metrics.RateLimitDropped.Inc()
			w.moveToIgnored(path, "rate_limit_exceeded")
			return
		}

		// Submit to worker pool
		w.pool.Submit(path)
	}(event.Name, ctx)
}

// processFile processes a single file (called by worker pool)
func (w *Watcher) processFile(ctx context.Context, path string) error {
	ctx, span := w.tracer.Start(ctx, "processFile")
	defer span.End()

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		metrics.FileProcessingDuration.Observe(duration.Seconds())
	}()

	span.SetAttributes(attribute.String("file.path", path))

	w.cfg.Logger.Info("Processing file", "path", path)

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		w.cfg.Logger.Error("Failed to stat file", "path", path, "error", err)
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Check file size
	size := info.Size()
	metrics.FileSizeBytes.Observe(float64(size))

	if size < w.cfg.MinFileSize {
		w.cfg.Logger.Warn("File too small", "path", path, "size", size, "min", w.cfg.MinFileSize)
		w.moveToIgnored(path, "file_too_small")
		metrics.FilesRejected.Inc()
		return nil
	}

	if size > w.cfg.MaxFileSize {
		w.cfg.Logger.Warn("File too large", "path", path, "size", size, "max", w.cfg.MaxFileSize)
		w.moveToIgnored(path, "file_too_large")
		metrics.FilesRejected.Inc()
		return nil
	}

	// Check if file is a ZIP and extract it
	if IsZipFile(path) {
		w.cfg.Logger.Info("ZIP file detected, extracting", "path", path)

		// Extract to processing directory
		processingDir := filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Processing)
		extractedFiles, err := ExtractZip(path, processingDir)
		if err != nil {
			w.cfg.Logger.Error("Failed to extract ZIP", "path", path, "error", err)
			w.moveToFailed(path, "zip_extraction_failed")
			metrics.WatcherErrors.Inc()
			return fmt.Errorf("failed to extract ZIP: %w", err)
		}

		w.cfg.Logger.Info("ZIP extracted successfully",
			"path", path,
			"files_extracted", len(extractedFiles))

		// Delete the ZIP file after successful extraction
		if err := os.Remove(path); err != nil {
			w.cfg.Logger.Warn("Failed to delete ZIP after extraction", "path", path, "error", err)
		} else {
			w.cfg.Logger.Info("ZIP file deleted after extraction", "path", path)
		}

		// Return early - extracted files will be processed by fsnotify
		return nil
	}

	// Calculate file hash
	hash, err := w.calculateHash(path)
	if err != nil {
		w.cfg.Logger.Error("Failed to calculate hash", "path", path, "error", err)
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	span.SetAttributes(attribute.String("file.hash", hash))

	// Check if already processed (idempotency)
	processed, err := w.cfg.Storage.IsProcessed(ctx, hash)
	if err != nil {
		w.cfg.Logger.Error("Failed to check if processed", "hash", hash, "error", err)
		metrics.StorageErrors.Inc()
	}

	if processed {
		w.cfg.Logger.Info("File already processed (duplicate)", "path", path, "hash", hash)
		w.moveToIgnored(path, "duplicate")
		metrics.FilesDuplicated.Inc()
		return nil
	}

	// Try to acquire lock (distributed locking)
	lock, err := w.cfg.Storage.GetLock(ctx, hash)
	if err != nil {
		w.cfg.Logger.Warn("Failed to acquire lock (another worker processing?)",
			"hash", hash, "error", err)
		metrics.FilesDuplicated.Inc()
		return nil // Not an error, just skip
	}
	defer lock.Release(ctx)

	// Move to processing directory
	processingPath, err := w.moveToProcessing(path)
	if err != nil {
		w.cfg.Logger.Error("Failed to move to processing", "path", path, "error", err)
		return fmt.Errorf("failed to move to processing: %w", err)
	}

	// Mark as enqueued
	if err := w.cfg.Storage.MarkEnqueued(ctx, hash, processingPath); err != nil {
		w.cfg.Logger.Error("Failed to mark as enqueued", "hash", hash, "error", err)
		metrics.StorageErrors.Inc()
	}

	// Create message
	msg := &queue.Message{
		ID:        hash,
		Path:      processingPath,
		Filename:  filepath.Base(path),
		Kind:      w.getFileKind(path),
		Size:      size,
		Hash:      hash,
		Timestamp: time.Now(),
	}

	// Publish to queue
	retryCfg := DefaultRetryConfig()

	// Wrap Retry with Circuit Breaker
	err = w.cb.Call(func() error {
		return Retry(ctx, retryCfg, func() error {
			return w.cfg.Queue.Publish(ctx, msg)
		})
	})

	if err != nil {
		w.cfg.Logger.Error("Failed to publish to queue after retries", "path", path, "error", err)

		// Move to failed directory
		w.moveToFailed(processingPath, fmt.Sprintf("queue_error: %v", err))

		// Mark as failed
		if err := w.cfg.Storage.MarkFailed(ctx, hash, err.Error()); err != nil {
			w.cfg.Logger.Error("Failed to mark as failed", "hash", hash, "error", err)
		}

		metrics.QueueErrors.Inc()
		return fmt.Errorf("failed to publish to queue: %w", err)
	}

	// Update metrics
	metrics.FilesSent.Inc()
	metrics.FilesProcessed.Inc()

	w.cfg.Logger.Info("File successfully enqueued",
		"path", path,
		"hash", hash,
		"size", size,
		"queue", msg.Kind)

	metrics.FilesSent.Inc()

	return nil
}

// ═══════════════════════════════════════════════════════════
//  HELPER FUNCTIONS - DIRECTORY OPERATIONS
// ═══════════════════════════════════════════════════════════

// createDirectories creates all required directories
func (w *Watcher) createDirectories() error {
	dirs := []string{
		w.cfg.WorkingDir,
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Processing),
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Processed),
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Failed),
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Ignored),
		filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Tmp),
	}

	// Add watch paths
	dirs = append(dirs, w.cfg.Paths...)

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// scanAndWatch performs initial scan and adds watchers recursively
func (w *Watcher) scanAndWatch(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			w.cfg.Logger.Error("Error accessing path during scan",
				"path", path,
				"error", err)
			return nil // Continue walking
		}

		if info.IsDir() {
			// Add directory to watcher
			if err := w.fsWatcher.Add(path); err != nil {
				w.cfg.Logger.Error("Failed to add directory to watcher",
					"path", path,
					"error", err)
			} else {
				w.cfg.Logger.Debug("Directory added to watcher", "path", path)
			}
		} else {
			// Process existing file
			if w.matchesPatterns(path) {
				w.cfg.Logger.Info("Processing existing file from scan", "path", path)
				w.pool.Submit(path)
			}
		}

		return nil
	})
}

// moveFile moves a file, handling cross-device link errors
func moveFile(src, dst string) error {
	// Try rename first (fast)
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// If cross-device, copy + delete
	if strings.Contains(err.Error(), "cross-device") || strings.Contains(err.Error(), "invalid cross-device link") {
		// Copy file
		srcFile, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("failed to open source: %w", err)
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("failed to create destination: %w", err)
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy: %w", err)
		}

		// Sync to disk
		if err := dstFile.Sync(); err != nil {
			return fmt.Errorf("failed to sync: %w", err)
		}

		// Delete source
		if err := os.Remove(src); err != nil {
			return fmt.Errorf("failed to remove source: %w", err)
		}

		return nil
	}

	return err
}

// moveToProcessing moves file to processing directory
func (w *Watcher) moveToProcessing(path string) (string, error) {
	filename := filepath.Base(path)
	destPath := filepath.Join(
		w.cfg.WorkingDir,
		w.cfg.SubDirs.Processing,
		filename,
	)

	if err := moveFile(path, destPath); err != nil {
		return "", fmt.Errorf("failed to move file: %w", err)
	}

	w.cfg.Logger.Debug("File moved to processing", "from", path, "to", destPath)

	return destPath, nil
}

// moveToFailed moves file to failed directory
func (w *Watcher) moveToFailed(path, reason string) {
	filename := filepath.Base(path)
	destPath := filepath.Join(
		w.cfg.WorkingDir,
		w.cfg.SubDirs.Failed,
		filename,
	)

	if err := os.Rename(path, destPath); err != nil {
		w.cfg.Logger.Error("Failed to move file to failed",
			"src", path,
			"dest", destPath,
			"error", err)
		return
	}

	w.cfg.Logger.Warn("File moved to failed",
		"path", destPath,
		"reason", reason)
}

// reconcileOrphans moves files from processing back to incoming to be re-processed
func (w *Watcher) reconcileOrphans() error {
	processingDir := filepath.Join(w.cfg.WorkingDir, w.cfg.SubDirs.Processing)

	// Check if processing directory exists
	if _, err := os.Stat(processingDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(processingDir)
	if err != nil {
		return fmt.Errorf("failed to read processing directory: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	w.cfg.Logger.Info("Found orphan files in processing directory", "count", len(entries))

	// Use the first watch path as the destination (incoming)
	if len(w.cfg.Paths) == 0 {
		return fmt.Errorf("no watch paths configured")
	}
	incomingDir := w.cfg.Paths[0]

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(processingDir, entry.Name())
		destPath := filepath.Join(incomingDir, entry.Name())

		w.cfg.Logger.Info("Reconciling orphan file", "file", entry.Name())

		if err := os.Rename(srcPath, destPath); err != nil {
			w.cfg.Logger.Error("Failed to move orphan file back to incoming",
				"file", entry.Name(),
				"error", err)
			continue
		}
	}

	return nil
}

// moveToIgnored moves file to ignored directory
func (w *Watcher) moveToIgnored(path, reason string) {
	filename := filepath.Base(path)
	destPath := filepath.Join(
		w.cfg.WorkingDir,
		w.cfg.SubDirs.Ignored,
		filename,
	)

	if err := os.Rename(path, destPath); err != nil {
		w.cfg.Logger.Error("Failed to move file to ignored",
			"src", path,
			"dest", destPath,
			"error", err)
		return
	}

	w.cfg.Logger.Info("File moved to ignored",
		"path", destPath,
		"reason", reason)

	metrics.FilesIgnored.Inc()
}

// ═══════════════════════════════════════════════════════════
//  HELPER FUNCTIONS - FILE OPERATIONS
// ═══════════════════════════════════════════════════════════

// matchesPatterns checks if a file matches the configured patterns
func (w *Watcher) matchesPatterns(path string) bool {
	filename := filepath.Base(path)

	// Check exclude patterns first
	for _, pattern := range w.cfg.ExcludePatterns {
		if matchPattern(filename, pattern) {
			return false
		}
	}

	// Check include patterns
	if len(w.cfg.FilePatterns) == 0 {
		return true // No patterns means match all
	}

	for _, pattern := range w.cfg.FilePatterns {
		if matchPattern(filename, pattern) {
			return true
		}
	}

	return false
}

// matchPattern matches a filename against a pattern
func matchPattern(filename, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}

	// Extension matching (e.g., "*.xml")
	if strings.HasPrefix(pattern, "*.") {
		ext := strings.ToLower(filepath.Ext(filename))
		patternExt := strings.ToLower(pattern[1:])
		return ext == patternExt
	}

	// Exact match
	if pattern == filename {
		return true
	}

	// Try filepath.Match for glob patterns
	matched, _ := filepath.Match(pattern, filename)
	return matched
}

// calculateHash calculates SHA256 hash of a file
func (w *Watcher) calculateHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()

	// Include filename in hash to avoid collisions with identical content
	// This allows processing multiple identical files (e.g. during testing)
	filename := filepath.Base(path)
	if _, err := hash.Write([]byte(filename)); err != nil {
		return "", fmt.Errorf("failed to write filename to hash: %w", err)
	}

	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getFileKind returns the file kind based on extension
func (w *Watcher) getFileKind(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) > 0 {
		return ext[1:] // Remove leading dot
	}
	return "unknown"
}

// ═══════════════════════════════════════════════════════════
//  HELPER FUNCTIONS - VALIDATION
// ═══════════════════════════════════════════════════════════

// validateConfig validates the watcher configuration
func validateConfig(cfg *Config) error {
	if len(cfg.Paths) == 0 {
		return fmt.Errorf("at least one watch path is required")
	}

	if cfg.MaxWorkers <= 0 {
		return fmt.Errorf("max_workers must be greater than 0")
	}

	if cfg.MaxFilesPerSecond <= 0 {
		return fmt.Errorf("max_files_per_second must be greater than 0")
	}

	if cfg.WorkingDir == "" {
		return fmt.Errorf("working_dir is required")
	}

	if cfg.Queue == nil {
		return fmt.Errorf("queue is required")
	}

	if cfg.Storage == nil {
		return fmt.Errorf("storage is required")
	}

	if cfg.Logger == nil {
		return fmt.Errorf("logger is required")
	}

	// Set defaults for subdirectories if not provided
	if cfg.SubDirs.Processing == "" {
		cfg.SubDirs.Processing = "processing"
	}
	if cfg.SubDirs.Processed == "" {
		cfg.SubDirs.Processed = "processed"
	}
	if cfg.SubDirs.Failed == "" {
		cfg.SubDirs.Failed = "failed"
	}
	if cfg.SubDirs.Ignored == "" {
		cfg.SubDirs.Ignored = "ignored"
	}
	if cfg.SubDirs.Tmp == "" {
		cfg.SubDirs.Tmp = "tmp"
	}

	return nil
}
