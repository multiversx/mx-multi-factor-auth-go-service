package storage

import (
	"errors"

	"github.com/multiversx/multi-factor-auth-go-service/redis"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("storage")

const noExpirationTime = 0

var errKeyNotFound = errors.New("key not found")

// ErrNilRedisClientWrapper signals that a nil redis client wrapper has been provided
var ErrNilRedisClientWrapper = errors.New("nil redis client wrapper provided")

type redisStorerHandler struct {
	client redis.RedisClientWrapper
}

// NewRedisStorerHandler will create a new redis storer handler instance
func NewRedisStorerHandler(client redis.RedisClientWrapper) (*redisStorerHandler, error) {
	if client == nil {
		return nil, ErrNilRedisClientWrapper
	}

	return &redisStorerHandler{
		client: client,
	}, nil
}

func (r *redisStorerHandler) Put(key []byte, data []byte) error {
	_, err := r.client.Set(string(key), data, noExpirationTime)
	if err != nil {
		return err
	}

	return nil
}

func (r *redisStorerHandler) Get(key []byte) ([]byte, error) {
	val, err := r.client.Get(string(key))
	if err != nil {
		return nil, err
	}

	return []byte(val), nil
}

func (r *redisStorerHandler) Has(key []byte) error {
	num, err := r.client.Exists(string(key))
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
	num, err := r.client.Del(string(key))
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
