package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Files (Vectors)
	filesDetectedVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_files_detected_total",
		Help: "Total number of files detected",
	}, []string{})

	filesSentVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_files_sent_total",
		Help: "Total number of files sent to queue",
	}, []string{})

	filesProcessedVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_files_processed_total",
		Help: "Total number of files successfully processed",
	}, []string{})

	filesDuplicatedVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_files_duplicated_total",
		Help: "Total number of duplicated files (already processed)",
	}, []string{})

	filesRejectedVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_files_rejected_total",
		Help: "Total number of rejected files",
	}, []string{})

	filesIgnoredVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_files_ignored_total",
		Help: "Total number of ignored files",
	}, []string{})

	// Errors (Vectors)
	watcherErrorsVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_errors_total",
		Help: "Total number of watcher errors",
	}, []string{})

	queueErrorsVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_queue_errors_total",
		Help: "Total number of queue publishing errors",
	}, []string{})

	storageErrorsVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_storage_errors_total",
		Help: "Total number of storage errors",
	}, []string{})

	// Rate Limiting (Vectors)
	rateLimitWaitsVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_rate_limit_waits_total",
		Help: "Number of times rate limiter caused a wait",
	}, []string{})

	rateLimitDroppedVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_rate_limit_dropped_total",
		Help: "Number of files dropped due to rate limiting",
	}, []string{})

	// Cleanup (Vectors)
	emptyDirectoriesRemovedVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gordon_watcher_empty_directories_removed_total",
		Help: "Total number of empty directories removed",
	}, []string{})

	// Public Counters (initialized from Vectors)
	FilesDetected           = filesDetectedVec.WithLabelValues()
	FilesSent               = filesSentVec.WithLabelValues()
	FilesProcessed          = filesProcessedVec.WithLabelValues()
	FilesDuplicated         = filesDuplicatedVec.WithLabelValues()
	FilesRejected           = filesRejectedVec.WithLabelValues()
	FilesIgnored            = filesIgnoredVec.WithLabelValues()
	WatcherErrors           = watcherErrorsVec.WithLabelValues()
	QueueErrors             = queueErrorsVec.WithLabelValues()
	StorageErrors           = storageErrorsVec.WithLabelValues()
	RateLimitWaits          = rateLimitWaitsVec.WithLabelValues()
	RateLimitDropped        = rateLimitDroppedVec.WithLabelValues()
	EmptyDirectoriesRemoved = emptyDirectoriesRemovedVec.WithLabelValues()

	// Worker Pool (Gauges don't need Vec for reset, they have Set)
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
// Uses CounterVec.Reset() to clear metrics and re-initializes them
func Reset() {
	// Reset Vectors
	filesDetectedVec.Reset()
	filesSentVec.Reset()
	filesProcessedVec.Reset()
	filesDuplicatedVec.Reset()
	filesRejectedVec.Reset()
	filesIgnoredVec.Reset()
	watcherErrorsVec.Reset()
	queueErrorsVec.Reset()
	storageErrorsVec.Reset()
	rateLimitWaitsVec.Reset()
	rateLimitDroppedVec.Reset()
	emptyDirectoriesRemovedVec.Reset()

	// Re-initialize Public Counters
	FilesDetected = filesDetectedVec.WithLabelValues()
	FilesSent = filesSentVec.WithLabelValues()
	FilesProcessed = filesProcessedVec.WithLabelValues()
	FilesDuplicated = filesDuplicatedVec.WithLabelValues()
	FilesRejected = filesRejectedVec.WithLabelValues()
	FilesIgnored = filesIgnoredVec.WithLabelValues()
	WatcherErrors = watcherErrorsVec.WithLabelValues()
	QueueErrors = queueErrorsVec.WithLabelValues()
	StorageErrors = storageErrorsVec.WithLabelValues()
	RateLimitWaits = rateLimitWaitsVec.WithLabelValues()
	RateLimitDropped = rateLimitDroppedVec.WithLabelValues()
	EmptyDirectoriesRemoved = emptyDirectoriesRemovedVec.WithLabelValues()

	// Reset gauges to 0
	WorkerPoolQueueSize.Set(0)
	WorkerPoolActiveWorkers.Set(0)
	GoroutineCount.Set(0)
}
