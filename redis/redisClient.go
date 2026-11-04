package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	pongValue = "PONG"
)

// decrementCounterScript decrements the counter, with zero as lower bound
var decrementCounterScript = redis.NewScript(`
local value = tonumber(redis.call("GET", KEYS[1]))
if value == nil then
	return 0
end
if value < 0 then
	redis.call("SET", KEYS[1], 0, "KEEPTTL")
	return 0
end
if value == 0 then
	return 0
end
return redis.call("DECR", KEYS[1])
`)

// setAllIfNoneExistScript sets all keys with the provided ttl in milliseconds, only if none of them exists
var setAllIfNoneExistScript = redis.NewScript(`
for i = 1, #KEYS do
	if redis.call("EXISTS", KEYS[i]) == 1 then
		return 0
	end
end
for i = 1, #KEYS do
	redis.call("SET", KEYS[i], 1, "PX", ARGV[1])
end
return 1
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
	return decrementCounterScript.Run(ctx, r.client, []string{key}).Int64()
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

// SetAllIfNoneExist will atomically set all the specified keys with the specified ttl, only if none of them exists yet
func (r *redisClientWrapper) SetAllIfNoneExist(ctx context.Context, keys []string, ttl time.Duration) (bool, error) {
	err := checkSetAllIfNoneExistArgs(keys, ttl)
	if err != nil {
		return false, err
	}

	wasSet, err := setAllIfNoneExistScript.Run(ctx, r.client, keys, ttl.Milliseconds()).Int64()
	if err != nil {
		return false, err
	}

	return wasSet == 1, nil
}

func checkSetAllIfNoneExistArgs(keys []string, ttl time.Duration) error {
	if len(keys) == 0 {
		return ErrNoKeysProvided
	}
	if ttl.Milliseconds() <= 0 {
		return ErrInvalidTTL
	}

	uniqueKeys := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		_, exists := uniqueKeys[key]
		if exists {
			return ErrDuplicatedKeys
		}
		uniqueKeys[key] = struct{}{}
	}

	return nil
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
