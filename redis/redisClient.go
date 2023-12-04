package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	pongValue = "PONG"
)

// redisClientWrapper defines a wrapper over redis client
type redisClientWrapper struct {
	client redis.UniversalClient
}

// NewRedisClientWrapper will create a new redis client wrapper component
func NewRedisClientWrapper(client redis.UniversalClient) (*redisClientWrapper, error) {
	if client == nil {
		return nil, ErrNilRedisClient
	}

	return &redisClientWrapper{
		client: client,
	}, nil
}

// Increment will run increment for the value corresponding to the specified key
func (r *redisClientWrapper) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// SetExpireIfNotExisting will run expire for the specified key, setting the specified ttl, if the key does not have a ttl set
func (r *redisClientWrapper) SetExpireIfNotExisting(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.ExpireNX(ctx, key, ttl).Result()
}

// Delete will delete the specified key
func (r *redisClientWrapper) Delete(ctx context.Context, key string) error {
	nDeleted, err := r.client.Del(ctx, key).Result()
	if nDeleted == 0 {
		log.Warn("no key to remove", "key", key)
	}

	return err
}

// ExpireTime will return expire time for the specified key
func (r *redisClientWrapper) ExpireTime(ctx context.Context, key string) (time.Duration, error) {
	expTime, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if expTime == -1 {
		return 0, ErrNoExpirationTimeForKey
	}
	if expTime == -2 {
		return 0, ErrKeyNotExists
	}

	return expTime, nil
}

// IsConnected will check if redis client is connected
func (r *redisClientWrapper) IsConnected(ctx context.Context) bool {
	pong, err := r.client.Ping(ctx).Result()
	return err == nil && pong == pongValue
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *redisClientWrapper) IsInterfaceNil() bool {
	return r == nil
}
