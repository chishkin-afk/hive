package cache

import (
	"context"
	"fmt"

	"github.com/braunkc/hive-engage/config"
	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, cfg config.CacheConfig) (*redis.Client, error) {
	opts := redis.Options{
		Addr:        fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:    cfg.Password,
		DB:          cfg.DB,
		PoolSize:    cfg.PoolSize,
		DialTimeout: cfg.DialTimeout,
		MaxRetries:  cfg.MaxRetries,
	}

	client := redis.NewClient(&opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
