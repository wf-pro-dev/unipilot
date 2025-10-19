package redis

import (
	"fmt"

	"github.com/redis/go-redis/v9"

	"unipilot/internal/secrets"
	"unipilot/internal/server"
)

func NewRedisClient() (*redis.Client, error) {
	redisAddr, err := secrets.GetEnvVar("REDIS_ADDRESS")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis url: %w", err)
	}
	redisPass, err := secrets.GetEnvVar("REDIS_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis password: %w", err)
	}
	server.PrintLOG([]string{"REDIS", "INFO"}, fmt.Sprintf("Redis URL: %s", redisAddr))
	return redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	}), nil
}
