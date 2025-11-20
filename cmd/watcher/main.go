package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/fabyo/gordon-watcher/internal/config"
	"github.com/fabyo/gordon-watcher/internal/health"
	"github.com/fabyo/gordon-watcher/internal/logger"
	"github.com/fabyo/gordon-watcher/internal/metrics"
	"github.com/fabyo/gordon-watcher/internal/queue"
	"github.com/fabyo/gordon-watcher/internal/storage"
	"github.com/fabyo/gordon-watcher/internal/telemetry"
	"github.com/fabyo/gordon-watcher/internal/watcher"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Setup logger
	loggerCfg := logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
		Output: cfg.Logger.Output,
	}
	appLog := logger.New(loggerCfg)

	appLog.Info("Gordon Watcher starting",
		"version", version,
		"commit", commit,
		"buildDate", buildDate,
	)

	appLog.Info("Config loaded",
		"queue.enabled", cfg.Queue.Enabled,
		"queue.type", cfg.Queue.Type,
		"redis.enabled", cfg.Redis.Enabled,
	)

	// Initialize metrics
	metrics.Init()

	// Start daily metrics reset at midnight
	metrics.StartDailyReset()
	appLog.Info("Daily metrics reset scheduled for midnight")

	// Start metrics server
	go func() {
		appLog.Info("Starting metrics server", "addr", cfg.Metrics.Addr)
		if err := metrics.StartServer(cfg.Metrics.Addr); err != nil {
			appLog.Error("Metrics server failed", "error", err)
		}
	}()

	// Start health check server
	healthServer := health.NewServer(cfg.Health.Addr, appLog)
	go func() {
		appLog.Info("Starting health check server", "addr", cfg.Health.Addr)
		if err := healthServer.Start(); err != nil {
			appLog.Error("Health server failed", "error", err)
		}
	}()

	// Initialize telemetry
	if cfg.Telemetry.Enabled {
		tp, err := telemetry.InitTracer(telemetry.Config{
			ServiceName: cfg.Telemetry.ServiceName,
			Endpoint:    cfg.Telemetry.Endpoint,
			Environment: cfg.App.Environment,
		})
		if err != nil {
			appLog.Warn("Failed to initialize telemetry", "error", err)
		} else {
			defer func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := tp.Shutdown(ctx); err != nil {
					appLog.Error("Failed to shutdown tracer", "error", err)
				}
			}()
			appLog.Info("Telemetry initialized")
		}
	}

	// Initialize storage
	var store storage.Storage
	if cfg.Redis.Enabled {
		redisStore, err := storage.NewRedisStorage(storage.RedisConfig{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err != nil {
			appLog.Error("Failed to initialize Redis", "error", err)
			appLog.Info("Falling back to memory storage")
			store = storage.NewMemoryStorage()
		} else {
			store = redisStore
			appLog.Info("Redis storage initialized")
		}
	} else {
		store = storage.NewMemoryStorage()
		appLog.Info("Memory storage initialized")
	}

	defer func() {
		if err := store.Close(); err != nil {
			appLog.Error("Error closing storage", "error", err)
		}
	}()

	// Initialize queue
	var q queue.Queue
	if cfg.Queue.Enabled {
		rabbitQueue, err := queue.NewRabbitMQQueue(queue.RabbitMQConfig{
			URL:        cfg.Queue.RabbitMQ.URL,
			Exchange:   cfg.Queue.RabbitMQ.Exchange,
			QueueName:  cfg.Queue.RabbitMQ.QueueName,
			RoutingKey: cfg.Queue.RabbitMQ.RoutingKey,
			Durable:    cfg.Queue.RabbitMQ.Durable,
		}, appLog)
		if err != nil {
			appLog.Error("Failed to initialize RabbitMQ", "error", err)
			appLog.Info("Falling back to NoOp queue")
			q = queue.NewNoOpQueue(appLog)
		} else {
			q = rabbitQueue
			appLog.Info("RabbitMQ queue initialized")
		}
	} else {
		q = queue.NewNoOpQueue(appLog)
		appLog.Info("NoOp queue initialized (queue disabled)")
	}

	defer func() {
		if err := q.Close(); err != nil {
			appLog.Error("Error closing queue", "error", err)
		}
	}()

	// Create watcher
	w, err := watcher.New(watcher.Config{
		Paths:             cfg.Watcher.Paths,
		FilePatterns:      cfg.Watcher.FilePatterns,
		ExcludePatterns:   cfg.Watcher.ExcludePatterns,
		MinFileSize:       cfg.Watcher.MinFileSize,
		MaxFileSize:       cfg.Watcher.MaxFileSize,
		StableAttempts:    cfg.Watcher.StableAttempts,
		StableDelay:       time.Duration(cfg.Watcher.StableDelay),     // ✅ Converter aqui
		CleanupInterval:   time.Duration(cfg.Watcher.CleanupInterval), // ✅ Converter aqui
		MaxWorkers:        cfg.Watcher.MaxWorkers,
		MaxFilesPerSecond: cfg.Watcher.MaxFilesPerSecond,
		WorkerQueueSize:   cfg.Watcher.WorkerQueueSize,
		WorkingDir:        cfg.Watcher.WorkingDir,
		SubDirs: watcher.SubDirectories{
			Processing: cfg.Watcher.SubDirectories.Processing,
			Processed:  cfg.Watcher.SubDirectories.Processed,
			Failed:     cfg.Watcher.SubDirectories.Failed,
			Ignored:    cfg.Watcher.SubDirectories.Ignored,
			Tmp:        cfg.Watcher.SubDirectories.Tmp,
		},
		Queue:   q,
		Storage: store,
		Logger:  appLog,
	})

	if err != nil {
		appLog.Error("Failed to create watcher", "error", err)
		os.Exit(1)
	}

	// Start watcher
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		appLog.Error("Failed to start watcher", "error", err)
		os.Exit(1)
	}

	// Mark as ready
	healthServer.SetReady(true)

	appLog.Info("Gordon Watcher started successfully",
		"paths", cfg.Watcher.Paths,
		"workers", cfg.Watcher.MaxWorkers,
		"rateLimit", cfg.Watcher.MaxFilesPerSecond,
	)

	// Start goroutine monitoring
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics.GoroutineCount.Set(float64(runtime.NumGoroutine()))
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	appLog.Info("Shutdown signal received", "signal", sig)

	// Mark as not ready
	healthServer.SetReady(false)

	// Graceful shutdown
	appLog.Info("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := w.Stop(shutdownCtx); err != nil {
		appLog.Error("Error during shutdown", "error", err)
	}

	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		appLog.Error("Error shutting down health server", "error", err)
	}

	appLog.Info("Gordon Watcher stopped")
}
