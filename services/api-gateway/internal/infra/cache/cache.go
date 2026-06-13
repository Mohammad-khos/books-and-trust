package cache

import (
	"context"
	"time"
)

type CacheStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	LoadScript(ctx context.Context, script string) (string, error)
	EvalSHA(ctx context.Context, sha string, keys []string, args ...any) (any, error)
	Close() error
}
