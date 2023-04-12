package bucket

import (
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const initialIndexValue = 0

type mongodbIndexHandler struct {
	usersColl     mongodb.CollectionID
	mongodbClient mongodb.MongoDBClient
}

// NewMongoDBIndexHandler returns a new instance of a mongodb index handler
func NewMongoDBIndexHandler(mongoClient mongodb.MongoDBClient, collID mongodb.CollectionID) (*mongodbIndexHandler, error) {
	if check.IfNil(mongoClient) {
		return nil, core.ErrNilMongoDBClient
	}
	if collID == "" {
		return nil, fmt.Errorf("%w: empty collection name", core.ErrInvalidValue)
	}

	handler := &mongodbIndexHandler{
		mongodbClient: mongoClient,
		usersColl:     collID,
	}

	err := handler.mongodbClient.PutIndexIfNotExists(handler.usersColl, []byte(lastIndexKey), initialIndexValue)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

// AllocateBucketIndex allocates a new index and returns it
func (handler *mongodbIndexHandler) AllocateBucketIndex() (uint32, error) {
	return handler.mongodbClient.IncrementIndex(handler.usersColl, []byte(lastIndexKey))
}

// Put adds data to storer
func (handler *mongodbIndexHandler) Put(key, data []byte) error {
	return handler.mongodbClient.Put(handler.usersColl, key, data)
}

// Get returns the value for the key from storer
func (handler *mongodbIndexHandler) Get(key []byte) ([]byte, error) {
	return handler.mongodbClient.Get(handler.usersColl, key)
}

// Has returns true if the key exists in storer
func (handler *mongodbIndexHandler) Has(key []byte) error {
	return handler.mongodbClient.Has(handler.usersColl, key)
}

// GetLastIndex returns the last index that was allocated
func (handler *mongodbIndexHandler) GetLastIndex() (uint32, error) {
	return handler.mongodbClient.GetIndex(handler.usersColl, []byte(lastIndexKey))
}

// Close closes the internal bucket
func (handler *mongodbIndexHandler) Close() error {
	return handler.mongodbClient.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *mongodbIndexHandler) IsInterfaceNil() bool {
	return handler == nil
}
