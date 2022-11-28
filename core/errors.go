package core

import "errors"

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilPubkeyConverter signals that nil address converter was provided
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilKeyGenerator signals that a nil key generator was provided
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrInvalidNumberOfBuckets signals that an invalid number of buckets was provided
var ErrInvalidNumberOfBuckets = errors.New("invalid number of buckets")
