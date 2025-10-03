package redis

import (
	"context"

	"github.com/spartan-truongvi/redis_rate/v10"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit redis_rate.Limit) (*redis_rate.Result, error)
	Peek(ctx context.Context, key string, limit redis_rate.Limit) (*redis_rate.Result, error)
	Reset(ctx context.Context, key string) error
}

func NewRedisRateLimiter(rds Redis) RateLimiter {
	return redis_rate.NewLimiter(rds.GetUniversalClient())
}
