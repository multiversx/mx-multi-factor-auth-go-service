package redis

import (
	"github.com/go-redis/redis_rate/v10"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
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
	return NewRateLimiter(rateLimiterArgs)
}

func createRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	switch cfg.ConnectionType {
	case core.RedisInstanceConnType:
		opt, err := redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, err
		}

		return redis.NewClient(opt), nil
	case core.RedisSentinelConnType:
		return createFailoverClient(cfg)
	default:
		return nil, core.ErrInvalidRedisConnType
	}
}

// createFailoverClient will create a redis client for a redis setup with sentinel
func createFailoverClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    cfg.MasterName,
		SentinelAddrs: []string{cfg.SentinelUrl},
	})

	return client, nil
}
