package config

import (
	"time"
)

// SetDefaults sets default values for configuration
func SetDefaults(cfg *Config) {
	// App defaults
	if cfg.App.Name == "" {
		cfg.App.Name = "gordon-watcher"
	}
	if cfg.App.Version == "" {
		cfg.App.Version = "dev"
	}
	if cfg.App.Environment == "" {
		cfg.App.Environment = "development"
	}

	// Watcher defaults
	if len(cfg.Watcher.Paths) == 0 {
		cfg.Watcher.Paths = []string{"/opt/gordon-watcher/data/incoming"}
	}
	if len(cfg.Watcher.FilePatterns) == 0 {
		cfg.Watcher.FilePatterns = []string{"*.xml", "*.zip"}
	}
	if len(cfg.Watcher.ExcludePatterns) == 0 {
		cfg.Watcher.ExcludePatterns = []string{".*", "*.tmp"}
	}
	if cfg.Watcher.MinFileSize == 0 {
		cfg.Watcher.MinFileSize = 100
	}
	if cfg.Watcher.MaxFileSize == 0 {
		cfg.Watcher.MaxFileSize = 100 * 1024 * 1024
	}
	if cfg.Watcher.StableAttempts == 0 {
		cfg.Watcher.StableAttempts = 5
	}
	if cfg.Watcher.StableDelay == 0 {
		cfg.Watcher.StableDelay = int64(1 * time.Second)
	}
	if cfg.Watcher.CleanupInterval == 0 {
		cfg.Watcher.CleanupInterval = int64(5 * time.Minute)
	}
	if cfg.Watcher.MaxWorkers == 0 {
		cfg.Watcher.MaxWorkers = 10
	}
	if cfg.Watcher.MaxFilesPerSecond == 0 {
		cfg.Watcher.MaxFilesPerSecond = 100
	}
	if cfg.Watcher.WorkerQueueSize == 0 {
		cfg.Watcher.WorkerQueueSize = 10
	}
	if cfg.Watcher.WorkingDir == "" {
		cfg.Watcher.WorkingDir = "/opt/gordon-watcher/data"
	}

	// SubDirectories defaults
	if cfg.Watcher.SubDirectories.Processing == "" {
		cfg.Watcher.SubDirectories.Processing = "processing"
	}
	if cfg.Watcher.SubDirectories.Processed == "" {
		cfg.Watcher.SubDirectories.Processed = "processed"
	}
	if cfg.Watcher.SubDirectories.Failed == "" {
		cfg.Watcher.SubDirectories.Failed = "failed"
	}
	if cfg.Watcher.SubDirectories.Ignored == "" {
		cfg.Watcher.SubDirectories.Ignored = "ignored"
	}
	if cfg.Watcher.SubDirectories.Tmp == "" {
		cfg.Watcher.SubDirectories.Tmp = "tmp"
	}

	// Queue defaults
	if cfg.Queue.Type == "" {
		cfg.Queue.Type = "rabbitmq"
	}
	if cfg.Queue.RabbitMQ.Exchange == "" {
		cfg.Queue.RabbitMQ.Exchange = "nfe_exchange"
	}
	if cfg.Queue.RabbitMQ.QueueName == "" {
		cfg.Queue.RabbitMQ.QueueName = "nfe_queue"
	}
	if cfg.Queue.RabbitMQ.RoutingKey == "" {
		cfg.Queue.RabbitMQ.RoutingKey = "nfe.xml"
	}

	// Redis defaults
	if cfg.Redis.Addr == "" {
		cfg.Redis.Addr = "localhost:6379"
	}

	// Metrics defaults
	if cfg.Metrics.Addr == "" {
		cfg.Metrics.Addr = ":9100"
	}

	// Health defaults
	if cfg.Health.Addr == "" {
		cfg.Health.Addr = ":8081"
	}

	// Telemetry defaults
	if cfg.Telemetry.ServiceName == "" {
		cfg.Telemetry.ServiceName = "gordon-watcher"
	}

	// Logger defaults
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Logger.Format == "" {
		cfg.Logger.Format = "json"
	}
	if cfg.Logger.Output == "" {
		cfg.Logger.Output = "stdout"
	}
}
