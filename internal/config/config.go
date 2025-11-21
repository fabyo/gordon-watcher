package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Watcher   WatcherConfig   `mapstructure:"watcher"`
	Queue     QueueConfig     `mapstructure:"queue"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	Health    HealthConfig    `mapstructure:"health"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	Cleanup   CleanupConfig   `mapstructure:"cleanup"`
}

// AppConfig holds application settings
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// WatcherConfig holds watcher settings
type WatcherConfig struct {
	Paths             []string             `mapstructure:"paths"`
	FilePatterns      []string             `mapstructure:"file_patterns"`
	ExcludePatterns   []string             `mapstructure:"exclude_patterns"`
	MinFileSize       int64                `mapstructure:"min_file_size"`
	MaxFileSize       int64                `mapstructure:"max_file_size"`
	StableAttempts    int                  `mapstructure:"stable_attempts"`
	StableDelay       int64                `mapstructure:"stable_delay"`
	CleanupInterval   int64                `mapstructure:"cleanup_interval"`
	MaxWorkers        int                  `mapstructure:"max_workers"`
	MaxFilesPerSecond int                  `mapstructure:"max_files_per_second"`
	WorkerQueueSize   int                  `mapstructure:"worker_queue_size"`
	WorkingDir        string               `mapstructure:"working_dir"`
	SubDirectories    SubDirectoriesConfig `mapstructure:"sub_directories"`
}

// SubDirectoriesConfig holds subdirectory names
type SubDirectoriesConfig struct {
	Processing string `mapstructure:"processing"`
	Processed  string `mapstructure:"processed"`
	Failed     string `mapstructure:"failed"`
	Ignored    string `mapstructure:"ignored"`
	Tmp        string `mapstructure:"tmp"`
}

// QueueConfig holds queue settings
type QueueConfig struct {
	Enabled  bool           `mapstructure:"enabled"`
	Type     string         `mapstructure:"type"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

// RabbitMQConfig holds RabbitMQ settings
type RabbitMQConfig struct {
	URL        string `mapstructure:"url"`
	Exchange   string `mapstructure:"exchange"`
	QueueName  string `mapstructure:"queue_name"`
	RoutingKey string `mapstructure:"routing_key"`
	Durable    bool   `mapstructure:"durable"`
}

// RedisConfig holds Redis settings
type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// MetricsConfig holds metrics settings
type MetricsConfig struct {
	Addr string `mapstructure:"addr"`
}

// HealthConfig holds health check settings
type HealthConfig struct {
	Addr string `mapstructure:"addr"`
}

// TelemetryConfig holds telemetry settings
type TelemetryConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	ServiceName string `mapstructure:"service_name"`
	Endpoint    string `mapstructure:"endpoint"`
}

// LoggerConfig holds logger settings
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// CleanupConfig holds cleanup settings
type CleanupConfig struct {
	Enabled   bool            `mapstructure:"enabled"`
	Schedule  string          `mapstructure:"schedule"`
	Retention RetentionConfig `mapstructure:"retention"`
}

// RetentionConfig holds retention periods in days
type RetentionConfig struct {
	Processed int `mapstructure:"processed"` // 0 = never delete
	Failed    int `mapstructure:"failed"`
	Ignored   int `mapstructure:"ignored"`
	Tmp       int `mapstructure:"tmp"` // 0 = always clean
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Config paths (in order of priority)
	viper.AddConfigPath("/etc/gordon/watcher")
	viper.AddConfigPath("$HOME/.gordon/watcher")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Environment variables override
	viper.SetEnvPrefix("GORDON_WATCHER")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// ✅ BIND EXPLÍCITO - ESSENCIAL!
	_ = viper.BindEnv("queue.enabled")
	_ = viper.BindEnv("queue.type")
	_ = viper.BindEnv("queue.rabbitmq.url")
	_ = viper.BindEnv("queue.rabbitmq.exchange")
	_ = viper.BindEnv("queue.rabbitmq.queue_name")
	_ = viper.BindEnv("queue.rabbitmq.routing_key")
	_ = viper.BindEnv("queue.rabbitmq.durable")

	_ = viper.BindEnv("redis.enabled")
	_ = viper.BindEnv("redis.addr")
	_ = viper.BindEnv("redis.password")
	_ = viper.BindEnv("redis.db")

	_ = viper.BindEnv("watcher.paths")
	_ = viper.BindEnv("watcher.working_dir")
	_ = viper.BindEnv("watcher.max_workers")
	_ = viper.BindEnv("watcher.max_files_per_second")

	_ = viper.BindEnv("telemetry.enabled")
	_ = viper.BindEnv("telemetry.service_name")
	_ = viper.BindEnv("telemetry.endpoint")

	_ = viper.BindEnv("app.environment")
	_ = viper.BindEnv("logger.level")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults + env vars
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// ✅ DEBUG
	fmt.Printf("DEBUG - cfg.Queue.Enabled: %v (type: %T)\n", cfg.Queue.Enabled, cfg.Queue.Enabled)
	fmt.Printf("DEBUG - cfg.Queue.Type: %v\n", cfg.Queue.Type)
	fmt.Printf("DEBUG - cfg.Redis.Enabled: %v\n", cfg.Redis.Enabled)
	fmt.Printf("DEBUG - Viper Get queue.enabled: %v\n", viper.GetBool("queue.enabled"))
	fmt.Printf("DEBUG - ENV GORDON_WATCHER_QUEUE_ENABLED: %v\n", os.Getenv("GORDON_WATCHER_QUEUE_ENABLED"))

	// Set defaults
	SetDefaults(&cfg)

	// Override with environment variables for critical settings
	if envWorkingDir := os.Getenv("GORDON_WATCHER_WORKING_DIR"); envWorkingDir != "" {
		cfg.Watcher.WorkingDir = envWorkingDir
	}
	if envPaths := os.Getenv("GORDON_WATCHER_PATHS"); envPaths != "" {
		cfg.Watcher.Paths = strings.Split(envPaths, ",")
	}

	return &cfg, nil
}
