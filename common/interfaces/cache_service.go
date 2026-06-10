package interfaces

import (
	"context"
	"time"
)

type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Exists(ctx context.Context, key string) (int64, error)
	Delete(ctx context.Context, key string) error
	Set(ctx context.Context, key string, data string, ttl time.Duration) error
	Close() error
}
