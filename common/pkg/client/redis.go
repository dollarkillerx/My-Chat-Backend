package client

import (
	"context"
	"time"

	"github.com/my-chat/common/pkg/config"
	"github.com/redis/go-redis/v9"
)

// RedisClient 创建Redis客户端
func RedisClient(cfg config.RedisConfiguration) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
