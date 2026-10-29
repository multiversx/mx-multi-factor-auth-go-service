package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	pongValue = "PONG"
)

// decrementIfPositiveScript decrements the counter only if the key exists and holds a positive value,
// so that the counter never drops below zero and a missing key is not created as a persistent one
var decrementIfPositiveScript = redis.NewScript(`
local value = tonumber(redis.call("GET", KEYS[1]))
if value == nil then
	return 0
end
if value <= 0 then
	return value
end
return redis.call("DECR", KEYS[1])
`)

// decrementIfPositiveScript decrements the counter only if it is greater than zero
var decrementIfPositiveScript = redis.NewScript(`
local value = tonumber(redis.call("GET", KEYS[1]))
if value == nil then
	return 0
end
if value <= 0 then
	return value
end
return redis.call("DECR", KEYS[1])
`)

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

// Decrement will atomically decrement the value corresponding to the specified key, without going below zero.
// It will not create the key if it does not exist
func (r *redisClientWrapper) Decrement(ctx context.Context, key string) (int64, error) {
	return decrementIfPositiveScript.Run(ctx, r.client, []string{key}).Int64()
}

// SetExpire will run expire for the specified key, setting the specified ttl
func (r *redisClientWrapper) SetExpire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.Expire(ctx, key, ttl).Result()
}

// SetExpireIfNotExists will run expire for the specified key, setting the specified ttl, only if ttl is not set yet
func (r *redisClientWrapper) SetExpireIfNotExists(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.ExpireNX(ctx, key, ttl).Result()
}

// SetPersist will run persist for the specified key without specified ttl
func (r *redisClientWrapper) SetPersist(ctx context.Context, key string) (bool, error) {
	return r.client.Persist(ctx, key).Result()
}

// SetGreaterExpireTTL will run expire for the specified key, setting the specified ttl, only if the new ttl greater than the old one
func (r *redisClientWrapper) SetGreaterExpireTTL(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.ExpireGT(ctx, key, ttl).Result()
}

// ResetCounterAndKeepTTL will reset the failures counter for the specified key, but will keep its ttl
func (r *redisClientWrapper) ResetCounterAndKeepTTL(ctx context.Context, key string) error {
	_, err := r.client.SetArgs(ctx, key, 0, redis.SetArgs{
		KeepTTL: true,
	}).Result()

	return err
}

// ExpireTime will return expire time for the specified key
func (r *redisClientWrapper) ExpireTime(ctx context.Context, key string) (time.Duration, error) {
	expTime, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
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
