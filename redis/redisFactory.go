package redis

import (
	"context"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/redis/go-redis/v9"

	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
)

var log = logger.GetOrCreate("redis")

// CreateRedisRateLimiter will create a new redis rate limiter component
func CreateRedisRateLimiter(cfg config.RedisConfig, twoFactorCfg config.TwoFactorConfig) (RateLimiter, error) {
	client, err := createRedisClient(cfg)
	if err != nil {
		return nil, err
	}
	redisStorer, err := NewRedisClientWrapper(client)
	if err != nil {
		return nil, err
	}

	ok := redisStorer.IsConnected(context.Background())
	if !ok {
		return nil, ErrRedisConnectionFailed
	}

	rateLimiterArgs := ArgsRateLimiter{
		OperationTimeoutInSec: cfg.OperationTimeoutInSec,
		FreezeFailureConfig: FailureConfig{
			MaxFailures:      twoFactorCfg.MaxFailures,
			LimitPeriodInSec: twoFactorCfg.BackoffTimeInSeconds,
		},
		SecurityModeFailureConfig: FailureConfig{
			MaxFailures:      twoFactorCfg.SecurityModeMaxFailures,
			LimitPeriodInSec: twoFactorCfg.SecurityModeBackoffTimeInSeconds,
		},
		Storer: redisStorer,
	}
	return NewRateLimiter(rateLimiterArgs)
}

func createRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	switch core.RedisConnType(cfg.ConnectionType) {
	case core.RedisInstanceConnType:
		return createSimpleClient(cfg)
	case core.RedisSentinelConnType:
		return createSentinelClient(cfg)
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

// createSentinelClient will create a redis client for a redis setup with sentinel
func createSentinelClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    cfg.MasterName,
		SentinelAddrs: []string{cfg.SentinelUrl},
	})

	log.Debug("created redis sentinel connection type", "connection url", cfg.SentinelUrl, "master", cfg.MasterName)

	return client, nil
}
