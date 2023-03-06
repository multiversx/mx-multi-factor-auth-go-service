package storage

import (
	"errors"

	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
)

//var log = logger.GetOrCreate("storage")

//var errKeyNotFound = errors.New("key not found")

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

func (r *mongodbStorerHandler) Put(key []byte, data []byte) error {
	return r.client.Put(r.collection, key, data)
}

func (r *mongodbStorerHandler) Get(key []byte) ([]byte, error) {
	return r.client.Get(r.collection, key)
}

func (r *mongodbStorerHandler) Has(key []byte) error {
	return r.client.Has(r.collection, key)
}

func (r *mongodbStorerHandler) SearchFirst(key []byte) ([]byte, error) {
	return r.Get(key)
}

func (r *mongodbStorerHandler) Remove(key []byte) error {
	return r.client.Remove(r.collection, key)
}

func (r *mongodbStorerHandler) ClearCache() {
	log.Warn("ClearCache: NOT implemented")
}

func (r *mongodbStorerHandler) Close() error {
	return r.client.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *mongodbStorerHandler) IsInterfaceNil() bool {
	return r == nil
}
