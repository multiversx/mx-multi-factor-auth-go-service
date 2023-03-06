package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBClientWrapper defines what a mongodb client wrapper should do
type MongoDBClientWrapper interface {
	GetCollection(coll Collection) *mongo.Collection
	Put(coll Collection, key []byte, data []byte) error
	Get(coll Collection, key []byte) ([]byte, error)
	Has(coll Collection, key []byte) error
	Remove(coll Collection, key []byte) error
	Close() error
	IsInterfaceNil() bool
}
