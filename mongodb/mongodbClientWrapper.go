package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type mongoDBClientWrapper struct {
	client *mongo.Client
}

func newMongoDBClientWrapper(client *mongo.Client) *mongoDBClientWrapper {
	return &mongoDBClientWrapper{
		client: client,
	}
}

// Connect will try to connect the db client
func (m *mongoDBClientWrapper) Connect(ctx context.Context) error {
	return m.client.Connect(ctx)
}

// Disconnect will disconnect the db client
func (m *mongoDBClientWrapper) Disconnect(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// DBCollection will return the specified collection object
func (m *mongoDBClientWrapper) DBCollection(dbName string, coll string) MongoDBCollection {
	return m.client.Database(dbName).Collection(coll)
}

// DBCollection will return the specified collection object
func (m *mongoDBClientWrapper) StartSession() (MongoDBSession, error) {
	return m.client.StartSession()
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *mongoDBClientWrapper) IsInterfaceNil() bool {
	return m == nil
}
