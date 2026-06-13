package ratelimiter

import "context"

type RateLimiter interface {
	Allow(ctx context.Context, key string, maxTokens int, refillRate int) (bool, error)
}
