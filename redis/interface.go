package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient defines what redis client should do
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
	IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd
	Close() error
}

// RedisClientWrapper defines what redis client wrapper should do
type RedisClientWrapper interface {
	Set(key string, value interface{}, expiration time.Duration) (string, error)
	Get(key string) (string, error)
	Exists(keys string) (int64, error)
	IncrBy(key []byte, increment int64) (int64, error)
	Del(keys string) (int64, error)
	Close() error
	IsInterfaceNil() bool
}
