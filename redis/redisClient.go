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

// SetEntry will set a new entry if not existing
func (r *redisClientWrapper) SetEntryIfNotExisting(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, ttl).Result()
}

// Delete will delete the specified key
func (r *redisClientWrapper) Delete(ctx context.Context, key string) error {
	nDeleted, err := r.client.Del(ctx, key).Result()
	if nDeleted == 0 {
		log.Warn("no key to remove", "key", key)
	}

	return err
}

// DecrementWithExpireTime will run decrement on value specified by key
// and returns the value after decrement and key expiry time
func (r *redisClientWrapper) DecrementWithExpireTime(ctx context.Context, key string) (int64, time.Duration, error) {
	v, err := r.Decrement(ctx, key)
	if err != nil {
		return 0, 0, err
	}

	expTime, err := r.ExpireTime(ctx, key)
	if err != nil {
		return 0, 0, err
	}

	return v, expTime, nil
}

// Decrement will run decrement for the value corresponding to the specified key
func (r *redisClientWrapper) Decrement(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
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

// IsConnected will check if redis clinet is connected
func (r *redisClientWrapper) IsConnected(ctx context.Context) bool {
	pong, err := r.client.Ping(ctx).Result()
	return err == nil && pong == pongValue
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *redisClientWrapper) IsInterfaceNil() bool {
	return r == nil
}
