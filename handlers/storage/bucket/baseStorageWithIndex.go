package bucket

import (
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/core"
)

type baseStorageWithIndex struct {
	bucketIDProvider core.BucketIDProvider
	bucketHandlers   map[uint32]core.IndexHandler
}

// AllocateIndex returns a new index that was not used before
func (bswi *baseStorageWithIndex) AllocateIndex(address []byte) (uint32, error) {
	bucketID, baseIndex, err := bswi.getBucketIDAndBaseIndex(address)
	if err != nil {
		return 0, err
	}

	return bswi.getNextFinalIndex(baseIndex, bucketID), nil
}

// Count returns the number of elements in all buckets
func (bswi *baseStorageWithIndex) Count() (uint32, error) {
	count := uint32(0)
	for idx, bucket := range bswi.bucketHandlers {
		numOfUsersInBucket, err := bucket.GetLastIndex()
		if err != nil {
			log.Error("could not get last index", "error", err, "bucket", idx)
			return 0, err
		}
		count += numOfUsersInBucket
	}

	return count, nil
}

func (bswi *baseStorageWithIndex) getBucketIDAndBaseIndex(address []byte) (uint32, uint32, error) {
	bucket, bucketID, err := bswi.getBucketForKey(address)
	if err != nil {
		return 0, 0, err
	}

	index, err := bucket.AllocateBucketIndex()
	return bucketID, index, err
}

func (bswi *baseStorageWithIndex) getBucketForKey(key []byte) (core.IndexHandler, uint32, error) {
	bucketID := bswi.bucketIDProvider.GetBucketForAddress(key)
	bucket, found := bswi.bucketHandlers[bucketID]
	if !found {
		return nil, 0, fmt.Errorf("%w for key %s", core.ErrInvalidBucketID, string(key))
	}

	return bucket, bucketID, nil
}

func (bswi *baseStorageWithIndex) getNextFinalIndex(newIndex, bucketID uint32) uint32 {
	numBuckets := uint32(len(bswi.bucketHandlers))
	return indexMultiplier * (newIndex*numBuckets + bucketID)
}
