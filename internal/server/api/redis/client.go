package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"unipilot/internal/secrets"
	"unipilot/internal/server"
)

func NewRedisClient() (*redis.Client, error) {
	redisAddr, err := secrets.GetEnvVar("REDIS_ADDR")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis url: %w", err)
	}
	// redisPass, err := secrets.GetEnvVar("REDIS_PASSWORD")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get redis password: %w", err)
	// }

	server.LogDebug(context.Background(), "Redis URL: ",
		"redis_addr", redisAddr,
		"tags", []string{"REDIS", "INFO"},
	)

	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
		//Password: redisPass,
		DB: 0,
	}), nil
}
