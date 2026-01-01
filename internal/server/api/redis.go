package server

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"unipilot/internal/errors"
	"unipilot/internal/secrets"
	"unipilot/internal/services/cache"
)

var RedisClient *redis.Client
var CacheService *cache.Cache

func NewRedisClient() error {
	redisAddr, err := secrets.GetEnvVar("REDIS_ADDR")
	if err != nil {
		return errors.Wrap(err, errors.ConfigEnvVarNotFound, "cannot get redis addr")
	}
	redisPass, err := secrets.GetEnvVar("REDIS_PASSWORD")
	if err != nil {
		return errors.Wrap(err, errors.ConfigEnvVarNotFound, "cannot get redis password")
	}

	RedisClient = redis.NewClient(&redis.Options{
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

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		RedisClient.Close()
		return err
	}

	// Initialize cache service
	CacheService = cache.New(RedisClient)

	return nil
}

func CloseRedis() error {
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			return errors.Wrap(err, errors.RedisCloseFailed, "cannot close redis connection")
		}
	}

	return nil
}
