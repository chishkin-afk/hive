package redissdk

import (
	"context"
	"fmt"

	"github.com/chishkin-afk/hive/backend/internal/common/config"
	"github.com/redis/go-redis/v9"
)

func Connect(cfg *config.Cache) (*redis.Client, error) {
	client := redis.NewClient(getRedisOpts(cfg))
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("can't ping redis conn: %w", err)
	}

	return client, nil
}

func getRedisOpts(cfg *config.Cache) *redis.Options {
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
