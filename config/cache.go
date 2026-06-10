package config

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
}

func NewCacheService(cacheLink string) (*CacheService, error) {
	opt, err := redis.ParseURL(cacheLink)
	if err != nil {
		return &CacheService{}, err
	}
	opt.PoolSize = 20
	opt.MinIdleConns = 5

	return &CacheService{
		client: redis.NewClient(opt),
	}, nil
}

func (c *CacheService) Set(ctx context.Context, key string, data string, ttl time.Duration) error {
	err := c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *CacheService) Get(ctx context.Context, key string) (string, error) {
	res, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return res, nil
}

func (c *CacheService) Exists(ctx context.Context, key string) (int64, error) {
	res, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *CacheService) Close() error {
	return c.client.Close()
}
