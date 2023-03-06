package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBClientWrapper defines what a mongodb client wrapper should do
type MongoDBClientWrapper interface {
	GetCollection(coll string) *mongo.Collection
	Close(ctx context.Context) error
	IsInterfaceNil() bool
}
