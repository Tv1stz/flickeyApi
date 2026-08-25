// Package db provides Redis client management.
package db

import (
	"context"
	"fmt"

	"flickey/go-backend/config"

	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis client from the configured URL.
func NewRedis(cfg *config.Settings) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Verify connection.
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connecting to Redis: %w", err)
	}

	return client, nil
}
