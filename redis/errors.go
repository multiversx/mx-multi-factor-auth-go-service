package redis

import "errors"

// ErrNilRedisClient signals that a nil redis client has been provided
var ErrNilRedisClient = errors.New("nil redis client")

// ErrInvalidKeyPrefix signals that an invalid key prefix has been provided
var ErrInvalidKeyPrefix = errors.New("invalid key prefix")

// ErrNoExpirationTimeForKey signals that key has no expiration time
var ErrNoExpirationTimeForKey = errors.New("key has no expiration time")

// ErrKeyNotExists signals that key does not exist
var ErrKeyNotExists = errors.New("key does not exist")

// ErrNilRedisClientWrapper signals that a nil redis client component has been provided
var ErrNilRedisClientWrapper = errors.New("nil redis client wrapper")

// ErrRedisConnectionFailed signals that connection to redis failed
var ErrRedisConnectionFailed = errors.New("error connecting to redis")

// ErrNoKeysProvided signals that no keys have been provided
var ErrNoKeysProvided = errors.New("no keys provided")

// ErrInvalidTTL signals that an invalid ttl has been provided
var ErrInvalidTTL = errors.New("invalid ttl")

// ErrDuplicatedKeys signals that duplicated keys have been provided
var ErrDuplicatedKeys = errors.New("duplicated keys")
