package redis

import (
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/redis/go-redis/v9"
)

// CreateRedisRateLimiter will create a new redis rate limiter component
func CreateRedisRateLimiter(cfg config.RedisConfig) (RateLimiter, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)

	limiter, err := NewRateLimiter(client)
	if err != nil {
		return nil, err
	}

	return limiter, nil
}
