package bucket

import (
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("bucket")

const (
	indexMultiplier = 2
)

// ArgShardedStorageWithIndex is the DTO used to create a new instance of sharded storage with index
type ArgShardedStorageWithIndex struct {
	BucketIDProvider core.BucketIDProvider
	BucketHandlers   map[uint32]core.IndexHandler
}

type shardedStorageWithIndex struct {
	*baseStorageWithIndex
}

// NewShardedStorageWithIndex returns a new instance of sharded storage with index
func NewShardedStorageWithIndex(args ArgShardedStorageWithIndex) (*shardedStorageWithIndex, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &shardedStorageWithIndex{
		baseStorageWithIndex: &baseStorageWithIndex{
			bucketIDProvider: args.BucketIDProvider,
			bucketHandlers:   args.BucketHandlers,
		},
	}, nil
}

func checkArgs(args ArgShardedStorageWithIndex) error {
	if check.IfNil(args.BucketIDProvider) {
		return core.ErrNilBucketIDProvider
	}
	if len(args.BucketHandlers) == 0 {
		return core.ErrInvalidBucketHandlers
	}

	for id, bucketHandler := range args.BucketHandlers {
		if check.IfNil(bucketHandler) {
			return fmt.Errorf("%w for id %d", core.ErrNilBucketHandler, id)
		}
	}

	return nil
}

// Put adds data to the bucket where the key should be
func (sswi *shardedStorageWithIndex) Put(key, data []byte) error {
	bucket, _, err := sswi.getBucketForKey(key)
	if err != nil {
		return err
	}

	return bucket.Put(key, data)
}

// Get returns the value for the key from the bucket where the key should be
func (sswi *shardedStorageWithIndex) Get(key []byte) ([]byte, error) {
	bucket, _, err := sswi.getBucketForKey(key)
	if err != nil {
		return make([]byte, 0), err
	}

	return bucket.Get(key)
}

// Has returns true if the key exists in the bucket where the key should be
func (sswi *shardedStorageWithIndex) Has(key []byte) error {
	bucket, _, err := sswi.getBucketForKey(key)
	if err != nil {
		return err
	}

	return bucket.Has(key)
}

// Close closes the managed buckets
func (sswi *shardedStorageWithIndex) Close() error {
	var lastError error
	for idx, bucket := range sswi.bucketHandlers {
		errClose := bucket.Close()
		if errClose != nil {
			lastError = errClose
			log.Error("could not close bucket", "error", lastError, "index", idx)
		}
	}

	return lastError
}

// IsInterfaceNil returns true if there is no value under the interface
func (sswi *shardedStorageWithIndex) IsInterfaceNil() bool {
	return sswi == nil
}
