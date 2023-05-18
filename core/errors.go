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

// ErrNilBucket signals that a nil bucket has been provided
var ErrNilBucket = errors.New("nil bucket")

// ErrNilMongoDBClient signals that a nil mongodb client has been provided
var ErrNilMongoDBClient = errors.New("nil mongodb client")

// ErrNilStorer signals that a nil storer has been provided
var ErrNilStorer = errors.New("nil storer")

// ErrNilHttpClient signals that a nil http client has been provided
var ErrNilHttpClient = errors.New("nil http client")

// ErrEmptyData signals that empty data was received
var ErrEmptyData = errors.New("empty data")

// ErrNilFacadeHandler signals that a nil facade handler has been provided
var ErrNilFacadeHandler = errors.New("nil facade handler")

// ErrNilMetricsHandler signals that a nil metrics handler has been provided
var ErrNilMetricsHandler = errors.New("nil metrics handler")
