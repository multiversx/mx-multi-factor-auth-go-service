package bucket

import (
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
	*baseStorageWithIndex
	mongodbClient mongodb.MongoDBClient
}

// NewMongoStorageWithIndex returns a new instance of mongo storage with index
func NewMongoStorageWithIndex(args ArgMongoStorageWithIndex) (*mongoStorageWithIndex, error) {
	return &mongoStorageWithIndex{
		baseStorageWithIndex: &baseStorageWithIndex{
			bucketIDProvider: args.BucketIDProvider,
			bucketHandlers:   args.IndexHandlers,
		},
		mongodbClient: args.MongoDBClient,
	}, nil
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

// Close closes the managed buckets
func (mswi *mongoStorageWithIndex) Close() error {
	return mswi.mongodbClient.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mswi *mongoStorageWithIndex) IsInterfaceNil() bool {
	return mswi == nil
}
