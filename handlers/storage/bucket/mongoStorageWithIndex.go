package bucket

import (
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
)

// ArgMongoStorageWithIndex defines the arguments needed to create a mongo storage with index
type ArgMongoStorageWithIndex struct {
	MongoDBClient    mongodb.MongoDBClient
	BucketIDProvider core.BucketIDProvider
	IndexHandlers    map[uint32]core.IndexHandler
}

type mongoStorageWithIndex struct {
	mongodbClient    mongodb.MongoDBClient
	bucketIDProvider core.BucketIDProvider
	indexHandlers    map[uint32]core.IndexHandler
}

// NewMongoStorageWithIndex returns a new instance of mongo storage with index
func NewMongoStorageWithIndex(args ArgMongoStorageWithIndex) (*mongoStorageWithIndex, error) {
	return &mongoStorageWithIndex{
		bucketIDProvider: args.BucketIDProvider,
		indexHandlers:    args.IndexHandlers,
		mongodbClient:    args.MongoDBClient,
	}, nil
}

// AllocateIndex returns a new index that was not used before
func (mswi *mongoStorageWithIndex) AllocateIndex(address []byte) (uint32, error) {
	bucketID, baseIndex, err := mswi.getBucketIDAndBaseIndex(address)
	if err != nil {
		return 0, err
	}

	return mswi.getNextFinalIndex(baseIndex, bucketID), nil
}

// Put adds data to the bucket where the key should be
func (mswi *mongoStorageWithIndex) Put(key, data []byte) error {
	return mswi.mongodbClient.Put(mongodb.UsersCollectionID, key, data)
}

// Get returns the value for the key from the bucket where the key should be
func (mswi *mongoStorageWithIndex) Get(key []byte) ([]byte, error) {
	return mswi.mongodbClient.Get(mongodb.UsersCollectionID, key)
}

// Has returns true if the key exists in the bucket where the key should be
func (mswi *mongoStorageWithIndex) Has(key []byte) error {
	return mswi.mongodbClient.Has(mongodb.UsersCollectionID, key)
}

// Count returns the number of elements in all buckets
func (mswi *mongoStorageWithIndex) Count() (uint32, error) {
	count := uint32(0)
	for idx, bucket := range mswi.indexHandlers {
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
func (mswi *mongoStorageWithIndex) Close() error {
	return mswi.mongodbClient.Close()
}

func (mswi *mongoStorageWithIndex) getBucketIDAndBaseIndex(address []byte) (uint32, uint32, error) {
	bucket, bucketID, err := mswi.getBucketForKey(address)
	if err != nil {
		return 0, 0, err
	}

	index, err := bucket.AllocateBucketIndex()
	return bucketID, index, err
}

func (mswi *mongoStorageWithIndex) getBucketForKey(key []byte) (core.IndexHandler, uint32, error) {
	bucketID := mswi.bucketIDProvider.GetBucketForAddress(key)
	bucket, found := mswi.indexHandlers[bucketID]
	if !found {
		return nil, 0, fmt.Errorf("%w for key %s", core.ErrInvalidBucketID, string(key))
	}

	return bucket, bucketID, nil
}

func (mswi *mongoStorageWithIndex) getNextFinalIndex(newIndex, bucketID uint32) uint32 {
	numBuckets := uint32(len(mswi.indexHandlers))
	return indexMultiplier * (newIndex*numBuckets + bucketID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mswi *mongoStorageWithIndex) IsInterfaceNil() bool {
	return mswi == nil
}
