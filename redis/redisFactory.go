package redis

import (
	"github.com/go-redis/redis_rate/v10"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/redis/go-redis/v9"
)

// CreateRedisRateLimiter will create a new redis rate limiter component
func CreateRedisRateLimiter(cfg config.RedisConfig, twoFactorCfg config.TwoFactorConfig) (RateLimiter, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	redisLimiter := redis_rate.NewLimiter(client)

	rateLimiterArgs := ArgsRateLimiter{
		OperationTimeoutInSec: cfg.OperationTimeoutInSec,
		MaxFailures:           twoFactorCfg.MaxFailures,
		LimitPeriodInSec:      twoFactorCfg.BackoffTimeInSeconds,
		Limiter:               redisLimiter,
	}
	limiter, err := NewRateLimiter(rateLimiterArgs)
	if err != nil {
		return nil, err
	}

	return limiter, nil
}
