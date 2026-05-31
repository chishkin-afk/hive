package redisconnect

import (
	"context"
	"fmt"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

func Connect(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(getOptions(&cfg.Cache))
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to open connection with redis: %w", err)
	}

	return client, nil
}

func getOptions(cfg *config.Cache) *redis.Options {
	return &redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Redis.Host,
			cfg.Redis.Port,
		),
		Username: cfg.Redis.Auth.Username,
		Password: cfg.Redis.Auth.Password,
		DB:       cfg.Redis.Auth.DB,
	}
}
