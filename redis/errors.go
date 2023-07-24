package redis

import "errors"

// ErrNilRedisClient signals that a nil redis client has been provided
var ErrNilRedisClient = errors.New("nil redis client")

// ErrInvalidKeyPrefix signals that an invalid key prefix has been provided
var ErrInvalidKeyPrefix = errors.New("invalid key prefix")

// ErrKeyNotFound signals that key does not exist
var ErrKeyNotFound = errors.New("failed to get key from cache: key not found")

// ErrNilRedisClientWrapper signals that a nil redis client component has been provided
var ErrNilRedisClientWrapper = errors.New("nil redis client wrapper")

// ErrRedisConnectionFailed signals that connection to redis failed
var ErrRedisConnectionFailed = errors.New("error connecting to redis")
