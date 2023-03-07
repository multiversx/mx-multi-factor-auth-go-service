package storage

import (
	"errors"

	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
)

// ErrNilRedisClientWrapper signals that a nil mongodb client has been provided
var ErrNilMongoDBClient = errors.New("nil mongodb client provided")

type mongodbStorerHandler struct {
	client     mongodb.MongoDBClientWrapper
	collection mongodb.Collection
}

// NewMongoDBStorerHandler will create a new storer handler instance
func NewMongoDBStorerHandler(client mongodb.MongoDBClientWrapper, collection mongodb.Collection) (*mongodbStorerHandler, error) {
	if client == nil {
		return nil, ErrNilMongoDBClient
	}

	return &mongodbStorerHandler{
		client:     client,
		collection: collection,
	}, nil
}

// Put will set key value pair
func (msh *mongodbStorerHandler) Put(key []byte, data []byte) error {
	return msh.client.Put(msh.collection, key, data)
}

// Get will return the value for the provided key
func (msh *mongodbStorerHandler) Get(key []byte) ([]byte, error) {
	return msh.client.Get(msh.collection, key)
}

// Has will return true if the provided key exists in the database collection
func (msh *mongodbStorerHandler) Has(key []byte) error {
	return msh.client.Has(msh.collection, key)
}

// SearchFirst will return the provided key
func (msh *mongodbStorerHandler) SearchFirst(key []byte) ([]byte, error) {
	return msh.Get(key)
}

// Remove will remove the provided key from the database collection
func (msh *mongodbStorerHandler) Remove(key []byte) error {
	return msh.client.Remove(msh.collection, key)
}

// ClearCache is not implemented
func (msh *mongodbStorerHandler) ClearCache() {
	log.Warn("ClearCache: NOT implemented")
}

// Close will close the mongodb client
func (msh *mongodbStorerHandler) Close() error {
	return msh.client.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (msh *mongodbStorerHandler) IsInterfaceNil() bool {
	return msh == nil
}
