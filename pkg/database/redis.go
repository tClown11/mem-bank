package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"mem_bank/pkg/logger"
)

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisClient creates and tests a new Redis client connection
func NewRedisClient(config RedisConfig, logger logger.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close() // Clean up on failure
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established")
	return client, nil
}

// NewRedisClientWithOptions creates and tests a Redis client with custom options
// Useful for testing with specific configurations
func NewRedisClientWithOptions(opts *redis.Options, timeout time.Duration) (*redis.Client, error) {
	client := redis.NewClient(opts)

	// Test connection with custom timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close() // Clean up on failure
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
