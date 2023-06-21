package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisClientWrapper implements a Redis-based store using go-redis v8.
type redisClientWrapper struct {
	client redis.UniversalClient
	prefix string
}

func NewRedisClientWrapper(client redis.UniversalClient, keyPrefix string) (*redisClientWrapper, error) {
	return &redisClientWrapper{
		client: client,
		prefix: keyPrefix,
	}, nil
}

func (r *redisClientWrapper) GetWithTime(ctx context.Context, key string) (int64, time.Time, error) {
	key = r.prefix + key

	pipe := r.client.Pipeline()
	timeCmd := pipe.Time(ctx)
	getKeyCmd := pipe.Get(ctx, key)
	_, err := pipe.Exec(ctx)

	now, err := timeCmd.Result()
	if err != nil {
		return 0, now, err
	}

	v, err := getKeyCmd.Int64()
	if err == redis.Nil {
		return -1, now, nil
	} else if err != nil {
		return 0, now, err
	}

	return v, now, nil
}

func (r *redisClientWrapper) SetEntryWithTTL(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
	key = r.prefix + key

	updated, err := r.client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, err
	}

	return updated, nil
}

func (r *redisClientWrapper) Delete(ctx context.Context, key string) error {
	key = r.prefix + key
	n, err := r.client.Del(ctx, key).Result()
	if n == 0 {
		log.Error("no key to remove", "key", key)
	}

	return err
}

func (r *redisClientWrapper) Decrement(ctx context.Context, key string) (int64, error) {
	key = r.prefix + key
	n, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (r *redisClientWrapper) ExpireTime(ctx context.Context, key string) (time.Duration, error) {
	key = r.prefix + key
	n, err := r.client.PTTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if n == -1 {
		return 0, errors.New("key has no expiration time")
	}
	if n == -2 {
		return 0, errors.New("key does not exist")
	}

	return n, nil
}
