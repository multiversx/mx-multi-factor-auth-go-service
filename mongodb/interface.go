package mongodb

import (
	"context"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Session defines the behaviour of a mongodb session
type Session mongo.Session

// SessionContext defines the behaviour of a mongodb session context
type SessionContext mongo.SessionContext

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
	PutIndexIfNotExists(collID CollectionID, key []byte, index uint32) error
	IncrementIndex(collID CollectionID, key []byte) (uint32, error)
	ReadWriteWithCheck(
		collID CollectionID,
		key []byte,
		checker func(data interface{}) (interface{}, error),
	) error
	ReadWithTx(
		collID CollectionID,
		key []byte,
	) ([]byte, Session, SessionContext, error)
	WriteWithTx(
		collID CollectionID,
		key []byte,
		value []byte,
		session Session,
		sessionCtx SessionContext,
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
