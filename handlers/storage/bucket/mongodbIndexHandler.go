package bucket

import (
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const initialIndexValue = 0

type mongodbIndexHandler struct {
	indexColl     mongodb.CollectionID
	mongodbClient mongodb.MongoDBClient
}

// NewMongoDBIndexHandler returns a new instance of a mongodb index handler
func NewMongoDBIndexHandler(mongoClient mongodb.MongoDBClient, collectionName string) (*mongodbIndexHandler, error) {
	if check.IfNil(mongoClient) {
		return nil, core.ErrNilMongoDBClient
	}

	handler := &mongodbIndexHandler{
		mongodbClient: mongoClient,
		indexColl:     mongodb.CollectionID(collectionName),
	}

	err := handler.mongodbClient.PutIndexIfNotExists(handler.indexColl, []byte(lastIndexKey), initialIndexValue)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

// AllocateBucketIndex allocates a new index and returns it
func (handler *mongodbIndexHandler) AllocateBucketIndex() (uint32, error) {
	newIndex, err := handler.mongodbClient.IncrementIndex(handler.indexColl, []byte(lastIndexKey))
	if err != nil {
		return 0, err
	}

	return newIndex, nil
}

// Put adds data to storer
func (handler *mongodbIndexHandler) Put(key, data []byte) error {
	return handler.mongodbClient.Put(mongodb.UsersCollectionID, key, data)
}

// Get returns the value for the key from storer
func (handler *mongodbIndexHandler) Get(key []byte) ([]byte, error) {
	return handler.mongodbClient.Get(mongodb.UsersCollectionID, key)
}

// Has returns true if the key exists in storer
func (handler *mongodbIndexHandler) Has(key []byte) error {
	return handler.mongodbClient.Has(mongodb.UsersCollectionID, key)
}

// GetLastIndex returns the last index that was allocated
func (handler *mongodbIndexHandler) GetLastIndex() (uint32, error) {
	return handler.mongodbClient.GetIndex(handler.indexColl, []byte(lastIndexKey))
}

// Close closes the internal bucket
func (handler *mongodbIndexHandler) Close() error {
	return handler.mongodbClient.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *mongodbIndexHandler) IsInterfaceNil() bool {
	return handler == nil
}
