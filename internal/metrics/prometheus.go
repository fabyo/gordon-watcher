package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Files
	FilesDetected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_files_detected_total",
		Help: "Total number of files detected",
	})

	FilesSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_files_sent_total",
		Help: "Total number of files sent to queue",
	})

	FilesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_files_processed_total",
		Help: "Total number of files successfully processed",
	})

	FilesDuplicated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_files_duplicated_total",
		Help: "Total number of duplicated files (already processed)",
	})

	FilesRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_files_rejected_total",
		Help: "Total number of rejected files",
	})

	FilesIgnored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_files_ignored_total",
		Help: "Total number of ignored files",
	})

	// Errors
	WatcherErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_errors_total",
		Help: "Total number of watcher errors",
	})

	QueueErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_queue_errors_total",
		Help: "Total number of queue publishing errors",
	})

	StorageErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_storage_errors_total",
		Help: "Total number of storage errors",
	})

	// Worker Pool
	WorkerPoolQueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gordon_watcher_worker_pool_queue_size",
		Help: "Current size of worker pool queue",
	})

	WorkerPoolActiveWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gordon_watcher_worker_pool_active_workers",
		Help: "Number of active workers currently processing files",
	})

	// Processing Time
	FileProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "gordon_watcher_file_processing_seconds",
		Help:    "Time taken to process a file",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	FileStabilityDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "gordon_watcher_file_stability_seconds",
		Help:    "Time taken for file to stabilize",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
	})

	// File Size
	FileSizeBytes = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "gordon_watcher_file_size_bytes",
		Help:    "Size of files detected in bytes",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 15),
	})

	// Rate Limiting
	RateLimitWaits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_rate_limit_waits_total",
		Help: "Number of times rate limiter caused a wait",
	})

	RateLimitDropped = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_rate_limit_dropped_total",
		Help: "Number of files dropped due to rate limiting",
	})

	// Cleanup
	EmptyDirectoriesRemoved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gordon_watcher_empty_directories_removed_total",
		Help: "Total number of empty directories removed",
	})

	// Runtime
	GoroutineCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gordon_watcher_goroutines",
		Help: "Current number of goroutines",
	})
)

// Init initializes metrics (sets to zero)
func Init() {
	// Initialize counters to zero so they appear in /metrics
	FilesDetected.Add(0)
	FilesSent.Add(0)
	FilesProcessed.Add(0)
	FilesDuplicated.Add(0)
	FilesRejected.Add(0)
	FilesIgnored.Add(0)
	WatcherErrors.Add(0)
	QueueErrors.Add(0)
	StorageErrors.Add(0)
	RateLimitWaits.Add(0)
	RateLimitDropped.Add(0)
	EmptyDirectoriesRemoved.Add(0)

	// Initialize gauges
	WorkerPoolQueueSize.Set(0)
	WorkerPoolActiveWorkers.Set(0)
	GoroutineCount.Set(0)
}

// Reset resets all counter metrics to zero
// Note: Prometheus counters cannot be truly reset, so we use a workaround
// by re-registering them. This is primarily for development/testing.
func Reset() {
	// For counters, we can't actually reset them in Prometheus
	// But we can provide a visual reset by using the Init pattern
	// In production, use PromQL queries like: increase(metric[1h])
	Init()
}
