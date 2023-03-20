package mongodb

import (
	"context"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBClientWrapper defines the methods for mongo db client wrapper
type MongoDBClientWrapper interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	DBCollection(dbName string, collName string) MongoDBCollection
	StartSession() (mongo.Session, error)
	IsInterfaceNil() bool
}

// MongoDBCollection defines the methods for mongo db collection
type MongoDBCollection interface {
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

// MongoDBClient defines what a mongodb client should do
type MongoDBClient interface {
	Put(coll CollectionID, key []byte, data []byte) error
	Get(coll CollectionID, key []byte) ([]byte, error)
	Has(coll CollectionID, key []byte) error
	Remove(coll CollectionID, key []byte) error
	IncrementWithTransaction(collID CollectionID, key []byte) (uint32, error)
	ReadWriteWithCheck(
		collID CollectionID,
		key []byte,
		checker func(data interface{}) (interface{}, error),
	) error
	Close() error
	IsInterfaceNil() bool
}

// MongoDBUsersHandler defines the behaviour of a mongo users handler component
type MongoDBUsersHandler interface {
	MongoDBClient
	UpdateTimestamp(collID CollectionID, key []byte, interval int64) (int64, error)
	PutStruct(collID CollectionID, key []byte, data *core.OTPInfo) error
	GetStruct(collID CollectionID, key []byte) (*core.OTPInfo, error)
	HasStruct(collID CollectionID, key []byte) error
}

// MongoDBSession defines what a mongodb session should do
type MongoDBSession interface {
	StartTransaction(...*options.TransactionOptions) error
	AbortTransaction(context.Context) error
	CommitTransaction(context.Context) error
	WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) (interface{}, error),
		opts ...*options.TransactionOptions) (interface{}, error)
	EndSession(context.Context)
}
