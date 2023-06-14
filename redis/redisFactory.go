package redis

import (
	"github.com/go-redis/redis_rate/v10"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/redis/go-redis/v9"
)

var log = logger.GetOrCreate("redis")

// CreateRedisRateLimiter will create a new redis rate limiter component
func CreateRedisRateLimiter(cfg config.RedisConfig, twoFactorCfg config.TwoFactorConfig) (RateLimiter, error) {
	client, err := createRedisClient(cfg)
	if err != nil {
		return nil, err
	}
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
		return createSimpleClient(cfg)
	case core.RedisSentinelConnType:
		return createFailoverClient(cfg)
	default:
		return nil, core.ErrInvalidRedisConnType
	}
}

func createSimpleClient(cfg config.RedisConfig) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	log.Debug("created redis instance connection type", "connection url", cfg.URL)

	return client, nil
}

// createFailoverClient will create a redis client for a redis setup with sentinel
func createFailoverClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    cfg.MasterName,
		SentinelAddrs: []string{cfg.SentinelUrl},
	})

	log.Debug("created redis sentinel connection type", "connection url", cfg.SentinelUrl, "master", cfg.MasterName)

	return client, nil
}
