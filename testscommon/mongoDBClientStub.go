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
	StartSessionCalled func() (mongodb.MongoDBSession, error)
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

// StartSession -
func (m *MongoDBClientWrapperStub) StartSession() (mongodb.MongoDBSession, error) {
	if m.StartSessionCalled != nil {
		return m.StartSessionCalled()
	}

	return nil, nil
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

// MongoDBSessionStub -
type MongoDBSessionStub struct {
	WithTransactionCalled func(ctx context.Context, fn func(sessCtx mongo.SessionContext) (interface{}, error), opts ...*options.TransactionOptions) (interface{}, error)
	EndSessionCalled      func(_ context.Context)
}

// WithTransaction -
func (m *MongoDBSessionStub) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) (interface{}, error), opts ...*options.TransactionOptions) (interface{}, error) {
	if m.WithTransactionCalled != nil {
		return m.WithTransactionCalled(ctx, fn)
	}

	return nil, nil
}

// EndSession -
func (m *MongoDBSessionStub) EndSession(ctx context.Context) {
	if m.EndSessionCalled != nil {
		m.EndSessionCalled(ctx)
	}
}
