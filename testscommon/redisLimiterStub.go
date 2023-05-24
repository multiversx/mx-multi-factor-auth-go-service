package testscommon

import (
	"context"

	"github.com/go-redis/redis_rate/v10"
)

// RedisLimiterStub -
type RedisLimiterStub struct {
	AllowCalled func(ctx context.Context, key string, limit redis_rate.Limit) (*redis_rate.Result, error)
	ResetCalled func(ctx context.Context, key string) error
}

// Allow -
func (r *RedisLimiterStub) Allow(ctx context.Context, key string, limit redis_rate.Limit) (*redis_rate.Result, error) {
	if r.AllowCalled != nil {
		return r.AllowCalled(ctx, key, limit)
	}

	return nil, nil
}

// Reset -
func (r *RedisLimiterStub) Reset(ctx context.Context, key string) error {
	if r.ResetCalled != nil {
		return r.ResetCalled(ctx, key)
	}

	return nil
}
