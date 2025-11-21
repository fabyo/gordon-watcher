package config

import (
	"fmt"
)

// Validate validates the configuration
func Validate(cfg *Config) error {
	// App validation
	if cfg.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	// Watcher validation
	if len(cfg.Watcher.Paths) == 0 {
		return fmt.Errorf("watcher.paths must have at least one path")
	}

	if cfg.Watcher.MaxWorkers <= 0 {
		return fmt.Errorf("watcher.max_workers must be greater than 0")
	}

	if cfg.Watcher.MaxFilesPerSecond <= 0 {
		return fmt.Errorf("watcher.max_files_per_second must be greater than 0")
	}

	if cfg.Watcher.WorkingDir == "" {
		return fmt.Errorf("watcher.working_dir is required")
	}

	// Queue validation
	if cfg.Queue.Enabled {
		if cfg.Queue.Type == "" {
			return fmt.Errorf("queue.type is required when queue is enabled")
		}

		if cfg.Queue.Type == "rabbitmq" {
			if cfg.Queue.RabbitMQ.URL == "" {
				return fmt.Errorf("queue.rabbitmq.url is required")
			}
			if cfg.Queue.RabbitMQ.Exchange == "" {
				return fmt.Errorf("queue.rabbitmq.exchange is required")
			}
			if cfg.Queue.RabbitMQ.QueueName == "" {
				return fmt.Errorf("queue.rabbitmq.queue_name is required")
			}
		}
	}

	// Redis validation
	if cfg.Redis.Enabled {
		if cfg.Redis.Addr == "" {
			return fmt.Errorf("redis.addr is required when redis is enabled")
		}
	}

	return nil
}
