package ratelimiter

import (
	"books-and-trust/services/api-gateway/internal/infra/cache"
	"context"
	"fmt"
	"time"
	_ "embed"
)

//go:embed token_bucket.lua
var tokenBucketLuaScript string

type TokenBucketLimiter struct {
	cacheStore cache.CacheStore
	scriptSHA  string
}

func NewTokenBucketLimiter(cacheStore cache.CacheStore) (*TokenBucketLimiter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sha, err := cacheStore.LoadScript(ctx, tokenBucketLuaScript)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize token bucket script: %w", err)
	}

	return &TokenBucketLimiter{
		cacheStore: cacheStore,
		scriptSHA:  sha,
	}, nil
}

func (t *TokenBucketLimiter) Allow(ctx context.Context, key string, maxTokens int, refillRate int) (bool, error) {
	now := time.Now().Unix() 

	result, err := t.cacheStore.EvalSHA(ctx, t.scriptSHA, []string{key}, maxTokens, refillRate, now)
	if err != nil {
		return false, fmt.Errorf("error evaluating rate limit script: %w", err)
	}

	val, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type from redis, got %T", result)
	}

	return val == 1, nil
}