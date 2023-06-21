package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNoExpirationTimeForKey signals that key has no expiration time
var ErrNoExpirationTimeForKey = errors.New("key has no expiration time")

// ErrKeyNotExists signals that key does not exist
var ErrKeyNotExists = errors.New("key does not exist")

// redisClientWrapper implements a Redis-based store using go-redis v8.
type redisClientWrapper struct {
	client redis.UniversalClient
	prefix string
}

// NewRedisClientWrapper will create a new redis client wrapper component
func NewRedisClientWrapper(client redis.UniversalClient, keyPrefix string) (*redisClientWrapper, error) {
	return &redisClientWrapper{
		client: client,
		prefix: keyPrefix,
	}, nil
}

// SetEntry will set a new entry if not existing
func (r *redisClientWrapper) SetEntryIfNotExisting(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
	key = r.prefix + key

	updated, err := r.client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, err
	}

	return updated, nil
}

// Delete will delete the sepcified key
func (r *redisClientWrapper) Delete(ctx context.Context, key string) error {
	key = r.prefix + key
	nDeleted, err := r.client.Del(ctx, key).Result()
	if nDeleted == 0 {
		log.Warn("no key to remove", "key", key)
	}

	return err
}

// Decrement will run decrement for the value corresponding to the specified key
func (r *redisClientWrapper) Decrement(ctx context.Context, key string) (int64, error) {
	key = r.prefix + key

	return r.client.Decr(ctx, key).Result()
}

// ExpireTime will return expire time for the specified key
func (r *redisClientWrapper) ExpireTime(ctx context.Context, key string) (time.Duration, error) {
	key = r.prefix + key

	expTime, err := r.client.PTTL(ctx, key).Result()
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

// IsInterfaceNil returns true if there is no value under the interface
func (r *redisClientWrapper) IsInterfaceNil() bool {
	return r == nil
}
