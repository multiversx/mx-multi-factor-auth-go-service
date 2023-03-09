package testscommon

import (
	"context"

	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBClientWrapperStub -
type MongoDBClientWrapperStub struct {
	DBCollectionCalled func(dbName string, collName string) mongodb.MongoDBCollection
	ConnectCalled      func(ctx context.Context) error
	DisconnectCalled   func(ctx context.Context) error
}

// DBCollection -
func (m *MongoDBClientWrapperStub) DBCollection(dbName string, collName string) mongodb.MongoDBCollection {
	if m.DBCollectionCalled != nil {
		return m.DBCollectionCalled(dbName, collName)
	}

	return &MongoDBCollectionStub{}
}

// Connect -
func (m *MongoDBClientWrapperStub) Connect(ctx context.Context) error {
	if m.ConnectCalled != nil {
		return m.ConnectCalled(ctx)
	}

	return nil
}

// Disconnect -
func (m *MongoDBClientWrapperStub) Disconnect(ctx context.Context) error {
	if m.DisconnectCalled != nil {
		return m.DisconnectCalled(ctx)
	}

	return nil
}

// IsInterfaceNil -
func (m *MongoDBClientWrapperStub) IsInterfaceNil() bool {
	return m == nil
}

// MongoDBCollectionStub -
type MongoDBCollectionStub struct {
	UpdateOneCalled func(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	FindOneCalled   func(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult

	DeleteOneCalled func(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

// UpdateOne -
func (m *MongoDBCollectionStub) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if m.UpdateOneCalled != nil {
		return m.UpdateOneCalled(ctx, filter, update)
	}

	return nil, nil
}

// FindOne -
func (m *MongoDBCollectionStub) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	if m.FindOneCalled != nil {
		return m.FindOneCalled(ctx, filter)
	}

	return nil
}

// DeleteOne -
func (m *MongoDBCollectionStub) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	if m.DeleteOneCalled != nil {
		return m.DeleteOneCalled(ctx, filter)
	}

	return nil, nil
}
