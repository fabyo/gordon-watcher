package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefixProcessed = "gordon:watcher:processed:"
	keyPrefixEnqueued  = "gordon:watcher:enqueued:"
	keyPrefixFailed    = "gordon:watcher:failed:"
	keyPrefixLock      = "gordon:watcher:lock:"

	defaultTTL   = 24 * time.Hour     // 24 hours
	lockTTL      = 30 * time.Second   // 30 seconds
	processedTTL = 7 * 24 * time.Hour // 7 days
)

// RedisConfig configures Redis connection
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// RedisStorage implements Storage using Redis
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage
func NewRedisStorage(cfg RedisConfig) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{
		client: client,
	}, nil
}

// IsProcessed checks if a file hash has been processed
func (s *RedisStorage) IsProcessed(ctx context.Context, hash string) (bool, error) {
	key := keyPrefixProcessed + hash

	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if processed: %w", err)
	}

	return exists > 0, nil
}

// MarkEnqueued marks a file as enqueued
func (s *RedisStorage) MarkEnqueued(ctx context.Context, hash, path string) error {
	key := keyPrefixEnqueued + hash

	err := s.client.Set(ctx, key, path, defaultTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to mark as enqueued: %w", err)
	}

	return nil
}

// MarkProcessed marks a file as processed
func (s *RedisStorage) MarkProcessed(ctx context.Context, hash string) error {
	key := keyPrefixProcessed + hash

	err := s.client.Set(ctx, key, time.Now().Unix(), processedTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to mark as processed: %w", err)
	}

	// Remove from enqueued
	enqueuedKey := keyPrefixEnqueued + hash
	_ = s.client.Del(ctx, enqueuedKey).Err()

	return nil
}

// MarkFailed marks a file as failed
func (s *RedisStorage) MarkFailed(ctx context.Context, hash, reason string) error {
	key := keyPrefixFailed + hash

	data := fmt.Sprintf("%d:%s", time.Now().Unix(), reason)
	err := s.client.Set(ctx, key, data, processedTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to mark as failed: %w", err)
	}

	// Remove from enqueued
	enqueuedKey := keyPrefixEnqueued + hash
	_ = s.client.Del(ctx, enqueuedKey).Err()

	return nil
}

// GetLock acquires a distributed lock
func (s *RedisStorage) GetLock(ctx context.Context, hash string) (Lock, error) {
	key := keyPrefixLock + hash
	value := fmt.Sprintf("%d", time.Now().UnixNano())

	// Try to acquire lock
	ok, err := s.client.SetNX(ctx, key, value, lockTTL).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !ok {
		return nil, fmt.Errorf("lock already held")
	}

	return &redisLock{
		client: s.client,
		key:    key,
		value:  value,
	}, nil
}

// Close closes the Redis connection
func (s *RedisStorage) Close() error {
	return s.client.Close()
}

// redisLock implements Lock using Redis
type redisLock struct {
	client *redis.Client
	key    string
	value  string
}

// Release releases the lock
func (l *redisLock) Release(ctx context.Context) error {
	// Use Lua script to ensure we only delete our own lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	err := l.client.Eval(ctx, script, []string{l.key}, l.value).Err()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}
