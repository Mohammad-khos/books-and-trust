package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotFound = errors.New("key not found in cache")

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string, password string, db int) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

func (r *RedisStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

func (r *RedisStore) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

func (r *RedisStore) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

func (r *RedisStore) LoadScript(ctx context.Context, script string) (string, error) {
	sha, err := r.client.ScriptLoad(ctx, script).Result()
	if err != nil {
		return "", fmt.Errorf("failed to load lua script: %w", err)
	}
	return sha, nil
}

func (r *RedisStore) EvalSHA(ctx context.Context, sha string, keys []string, args ...any) (any, error) {
	result, err := r.client.EvalSha(ctx, sha, keys, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to execute evalsha: %w", err)
	}
	return result, nil
}

func (r *RedisStore) Close() error {
	return r.client.Close()
}
