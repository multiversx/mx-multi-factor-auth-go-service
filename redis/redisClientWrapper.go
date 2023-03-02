package redis

import (
	"context"
	"errors"
	"time"
)

// ErrNilRedisClient signals that a nil redis client has been provided
var ErrNilRedisClient = errors.New("nil redis client provided")

type redisClientWrapper struct {
	client RedisClient
	ctx    context.Context
}

// NewRedisClientWrapper will create a new redis client wrapper
func NewRedisClientWrapper(client RedisClient) (*redisClientWrapper, error) {
	if client == nil {
		return nil, ErrNilRedisClient
	}

	ctx := context.Background()

	return &redisClientWrapper{
		client: client,
		ctx:    ctx,
	}, nil
}

// Set will set key value pair
func (r *redisClientWrapper) Set(key string, value interface{}, expiration time.Duration) (string, error) {
	return r.client.Set(r.ctx, key, value, expiration).Result()
}

// Get will return the value for the provided key
func (r *redisClientWrapper) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// Exists return num of existing keys form provided list
func (r *redisClientWrapper) Exists(keys string) (int64, error) {
	return r.client.Exists(r.ctx, keys).Result()
}

// IncrBy will add increment to the value corresponding to provided key
// It will return the value after the increment
func (r *redisClientWrapper) IncrBy(key []byte, increment int64) (int64, error) {
	return r.client.IncrBy(r.ctx, string(key), increment).Result()
}

// Del will delete specified keys
func (r *redisClientWrapper) Del(keys string) (int64, error) {
	return r.client.Del(r.ctx, keys).Result()
}

// Close will close the redis connection
func (r *redisClientWrapper) Close() error {
	return r.client.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *redisClientWrapper) IsInterfaceNil() bool {
	return r == nil
}
