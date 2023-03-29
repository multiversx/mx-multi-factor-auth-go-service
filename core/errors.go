package core

import "errors"

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilKeyGenerator signals that a nil key generator was provided
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrInvalidNumberOfBuckets signals that an invalid number of buckets was provided
var ErrInvalidNumberOfBuckets = errors.New("invalid number of buckets")

// ErrNilBucketIDProvider signals that a nil bucket id provider was provided
var ErrNilBucketIDProvider = errors.New("nil bucket id provider")

// ErrInvalidBucketID signals that an invalid bucket id was provided
var ErrInvalidBucketID = errors.New("invalid bucket id")

// ErrInvalidBucketHandlers signals than an invalid bucket handlers was provided
var ErrInvalidBucketHandlers = errors.New("invalid bucket handlers")

// ErrNilBucketHandler signals that a nil bucket handler was provided
var ErrNilBucketHandler = errors.New("nil bucket handler")

// ErrNilBucket is raised when a nil bucket has been provided
var ErrNilBucket = errors.New("nil bucket")

// ErrNilMongoDBClient signals that a nil mongodb client has been provided
var ErrNilMongoDBClient = errors.New("nil mongodb client")

// ErrNilStorer signals that a nil storer has been provided
var ErrNilStorer = errors.New("nil storer")
