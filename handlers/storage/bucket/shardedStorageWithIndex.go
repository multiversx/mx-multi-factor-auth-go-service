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
	bucketIDProvider core.BucketIDProvider
	bucketHandlers   map[uint32]core.IndexHandler
}

// NewShardedStorageWithIndex returns a new instance of sharded storage with index
func NewShardedStorageWithIndex(args ArgShardedStorageWithIndex) (*shardedStorageWithIndex, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &shardedStorageWithIndex{
		bucketIDProvider: args.BucketIDProvider,
		bucketHandlers:   args.BucketHandlers,
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

// AllocateIndex returns a new index that was not used before
func (sswi *shardedStorageWithIndex) AllocateIndex(address []byte) (uint32, error) {
	bucketID, baseIndex, err := sswi.getBucketIDAndBaseIndex(address)
	if err != nil {
		return 0, err
	}

	return sswi.getNextFinalIndex(baseIndex, bucketID), nil
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

func (sswi *shardedStorageWithIndex) UpdateWithCheck(key []byte, fn func(data interface{}) (interface{}, error)) error {
	bucket, _, err := sswi.getBucketForKey(key)
	if err != nil {
		return err
	}

	return bucket.UpdateWithCheck(key, fn)
}

// Count returns the number of elements in all buckets
func (sswi *shardedStorageWithIndex) Count() (uint32, error) {
	count := uint32(0)
	for idx, bucket := range sswi.bucketHandlers {
		numOfUsersInBucket, err := bucket.GetLastIndex()
		if err != nil {
			log.Error("could not get last index", "error", err, "bucket", idx)
			return 0, err
		}
		count += numOfUsersInBucket
	}

	return count, nil
}

// Close closes the managed buckets
func (sswi *shardedStorageWithIndex) Close() error {
	var lastError error
	for idx, bucket := range sswi.bucketHandlers {
		errClose := bucket.Close()
		if errClose != nil {
			lastError = errClose
			log.Error("could not close bucket, error: %w, index: %idx", lastError, idx)
		}
	}

	return lastError
}

func (sswi *shardedStorageWithIndex) getBucketIDAndBaseIndex(address []byte) (uint32, uint32, error) {
	bucket, bucketID, err := sswi.getBucketForKey(address)
	if err != nil {
		return 0, 0, err
	}

	index, err := bucket.AllocateBucketIndex()
	return bucketID, index, err
}

func (sswi *shardedStorageWithIndex) getBucketForKey(key []byte) (core.IndexHandler, uint32, error) {
	bucketID := sswi.bucketIDProvider.GetBucketForAddress(key)
	bucket, found := sswi.bucketHandlers[bucketID]
	if !found {
		return nil, 0, fmt.Errorf("%w for key %s", core.ErrInvalidBucketID, string(key))
	}

	return bucket, bucketID, nil
}

func (sswi *shardedStorageWithIndex) getNextFinalIndex(newIndex, bucketID uint32) uint32 {
	numBuckets := uint32(len(sswi.bucketHandlers))
	return indexMultiplier * (newIndex*numBuckets + bucketID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sswi *shardedStorageWithIndex) IsInterfaceNil() bool {
	return sswi == nil
}
