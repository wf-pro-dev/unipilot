package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"unipilot/internal/secrets"
	"unipilot/internal/server"
)

func NewRedisClient() (*redis.Client, error) {
	redisAddr, err := secrets.GetEnvVar("REDIS_ADDR")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis url: %w", err)
	}
	redisPass, err := secrets.GetEnvVar("REDIS_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis password: %w", err)
	}

	server.LogDebug(context.Background(), "Redis URL: ",
		"redis_addr", redisAddr,
		"tags", []string{"REDIS", "INFO"},
	)

	client := redis.NewClient(&redis.Options{
		Addr:            redisAddr,
		Password:        redisPass,
		DB:              0,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        10,
		MinIdleConns:    5,
		MaxIdleConns:    10,
		PoolTimeout:     4 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 30 * time.Minute,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}
