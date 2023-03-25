package bucket

import (
	"sync"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const initialIndexValue = 1

type mongodbIndexHandler struct {
	storer        core.Storer
	mongodbClient mongodb.MongoDBClient
	mut           sync.RWMutex
}

// NewMongoDBIndexHandler returns a new instance of a bucket index handler
func NewMongoDBIndexHandler(storer core.Storer, mongoClient mongodb.MongoDBClient) (*mongodbIndexHandler, error) {
	if check.IfNil(storer) {
		return nil, core.ErrNilStorer
	}
	if check.IfNil(mongoClient) {
		return nil, core.ErrNilMongoDBClient
	}

	handler := &mongodbIndexHandler{
		storer:        storer,
		mongodbClient: mongoClient,
	}

	err := handler.mongodbClient.PutIndex(mongodb.UsersCollectionID, []byte(lastIndexKey), initialIndexValue)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

// AllocateBucketIndex allocates a new index and returns it
func (handler *mongodbIndexHandler) AllocateBucketIndex() (uint32, error) {
	handler.mut.Lock()
	defer handler.mut.Unlock()

	newIndex, err := handler.mongodbClient.IncrementIndex(mongodb.UsersCollectionID, []byte(lastIndexKey))
	if err != nil {
		return 0, err
	}

	return newIndex, nil
}

// Put adds data to storer
func (handler *mongodbIndexHandler) Put(key, data []byte) error {
	return handler.storer.Put(key, data)
}

// Get returns the value for the key from storer
func (handler *mongodbIndexHandler) Get(key []byte) ([]byte, error) {
	return handler.storer.Get(key)
}

// Has returns true if the key exists in storer
func (handler *mongodbIndexHandler) Has(key []byte) error {
	return handler.storer.Has(key)
}

// GetLastIndex returns the last index that was allocated
func (handler *mongodbIndexHandler) GetLastIndex() (uint32, error) {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	return getIndex(handler.storer)
}

// Close closes the internal bucket
func (handler *mongodbIndexHandler) Close() error {
	handler.mut.Lock()
	defer handler.mut.Unlock()

	return handler.storer.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *mongodbIndexHandler) IsInterfaceNil() bool {
	return handler == nil
}
