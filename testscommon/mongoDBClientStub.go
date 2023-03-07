package testscommon

import (
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBClientStub implemented mongodb client wraper interface
type MongoDBClientStub struct {
	GetCollectionCalled func(coll mongodb.Collection) *mongo.Collection
	PutCalled           func(coll mongodb.Collection, key []byte, data []byte) error
	GetCalled           func(coll mongodb.Collection, key []byte) ([]byte, error)
	HasCalled           func(coll mongodb.Collection, key []byte) error
	RemoveCalled        func(coll mongodb.Collection, key []byte) error
	CloseCalled         func() error
}

// GetCollection -
func (m *MongoDBClientStub) GetCollection(coll mongodb.Collection) *mongo.Collection {
	if m.GetCollectionCalled != nil {
		return m.GetCollectionCalled(coll)
	}

	return nil
}

// Put -
func (m *MongoDBClientStub) Put(coll mongodb.Collection, key []byte, data []byte) error {
	if m.PutCalled != nil {
		return m.PutCalled(coll, key, data)
	}

	return nil
}

// Get -
func (m *MongoDBClientStub) Get(coll mongodb.Collection, key []byte) ([]byte, error) {
	if m.GetCalled != nil {
		return m.GetCalled(coll, key)
	}

	return nil, nil
}

// Has -
func (m *MongoDBClientStub) Has(coll mongodb.Collection, key []byte) error {
	if m.HasCalled != nil {
		return m.HasCalled(coll, key)
	}

	return nil
}

// Remove -
func (m *MongoDBClientStub) Remove(coll mongodb.Collection, key []byte) error {
	if m.RemoveCalled != nil {
		return m.RemoveCalled(coll, key)
	}

	return nil
}

// Close -
func (m *MongoDBClientStub) Close() error {
	if m.CloseCalled != nil {
		return m.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (m *MongoDBClientStub) IsInterfaceNil() bool {
	return m == nil
}
