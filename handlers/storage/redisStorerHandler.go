package storage

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("storage")

const noExpirationTime = 0

var errKeyNotFound = errors.New("key not found")
var errNilRedisClient = errors.New("nil redis client provided")

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
}

type redisStorerHandler struct {
	client RedisClient
}

// NewRedisStorerHandler will create a new redis storer handler instance
func NewRedisStorerHandler(client RedisClient) (*redisStorerHandler, error) {
	if client == nil {
		return nil, errNilRedisClient
	}

	return &redisStorerHandler{
		client: client,
	}, nil
}

func (r *redisStorerHandler) Put(key []byte, data []byte) error {
	_, err := r.client.Set(context.Background(), string(key), data, noExpirationTime).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *redisStorerHandler) Get(key []byte) ([]byte, error) {
	val, err := r.client.Get(context.Background(), string(key)).Result()
	if err != nil {
		return nil, err
	}

	return []byte(val), nil
}

func (r *redisStorerHandler) Has(key []byte) error {
	num, err := r.client.Exists(context.Background(), string(key)).Result()
	if err != nil {
		return err
	}

	if num > 0 {
		return nil
	}

	return errKeyNotFound
}

func (r *redisStorerHandler) SearchFirst(key []byte) ([]byte, error) {
	return r.Get(key)
}

func (r *redisStorerHandler) Remove(key []byte) error {
	num, err := r.client.Del(context.Background(), string(key)).Result()
	if err != nil {
		return err
	}

	if num == 0 {
		log.Debug("no key to remove")
	}

	return nil
}

func (r *redisStorerHandler) ClearCache() {
	log.Warn("ClearCache: NOT implemented")
}

func (r *redisStorerHandler) Close() error {
	return r.client.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *redisStorerHandler) IsInterfaceNil() bool {
	return r == nil
}
