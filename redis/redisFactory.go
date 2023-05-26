package redis

import (
	"github.com/go-redis/redis_rate/v10"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/redis/go-redis/v9"
)

// CreateRedisRateLimiter will create a new redis rate limiter component
func CreateRedisRateLimiter(cfg config.RedisConfig, twoFactorCfg config.TwoFactorConfig) (RateLimiter, error) {
	opt, err := redis.ParseURL(cfg.Cacher.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	redisLimiter := redis_rate.NewLimiter(client)

	rateLimiterArgs := ArgsRateLimiter{
		OperationTimeoutInSec: cfg.Cacher.OperationTimeoutInSec,
		MaxFailures:           twoFactorCfg.MaxFailures,
		LimitPeriodInSec:      twoFactorCfg.BackoffTimeInSeconds,
		Limiter:               redisLimiter,
	}
	return NewRateLimiter(rateLimiterArgs)
}

func CreateRedisLocker(cfg config.RedisConfig) (Locker, error) {
	opt, err := redis.ParseURL(cfg.Locker.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return NewRedisLockerWrapper(rs, cfg.Locker.LockTimeExpiryInSec)
}
