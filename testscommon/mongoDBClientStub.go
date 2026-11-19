package testscommon

import (
	"github.com/multiversx/mx-multi-factor-auth-go-service/mongodb"
)

// MongoDBClientStub implemented mongodb client wraper interface
type MongoDBClientStub struct {
	PutCalled                  func(coll mongodb.CollectionID, key []byte, data []byte) error
	GetCalled                  func(coll mongodb.CollectionID, key []byte) ([]byte, error)
	GetWithVersionCalled       func(coll mongodb.CollectionID, key []byte) ([]byte, int64, error)
	InsertCalled               func(coll mongodb.CollectionID, key []byte, data []byte) error
	PutIfVersionCalled         func(coll mongodb.CollectionID, key []byte, data []byte, expectedVersion int64) error
	HasCalled                  func(coll mongodb.CollectionID, key []byte) error
	RemoveCalled               func(coll mongodb.CollectionID, key []byte) error
	GetIndexCalled             func(collID mongodb.CollectionID, key []byte) (uint32, error)
	IncrementIndexCalled       func(collID mongodb.CollectionID, key []byte) (uint32, error)
	PutIndexIfNotExistsCalled  func(collID mongodb.CollectionID, key []byte, index uint32) error
	GetAllCollectionsIDsCalled func() []mongodb.CollectionID
	CloseCalled                func() error
}

// Put -
func (m *MongoDBClientStub) Put(coll mongodb.CollectionID, key []byte, data []byte) error {
	if m.PutCalled != nil {
		return m.PutCalled(coll, key, data)
	}

	return nil
}

// Get -
func (m *MongoDBClientStub) Get(coll mongodb.CollectionID, key []byte) ([]byte, error) {
	if m.GetCalled != nil {
		return m.GetCalled(coll, key)
	}

	return nil, nil
}

// GetWithVersion -
func (m *MongoDBClientStub) GetWithVersion(coll mongodb.CollectionID, key []byte) ([]byte, int64, error) {
	if m.GetWithVersionCalled != nil {
		return m.GetWithVersionCalled(coll, key)
	}

	return nil, 0, nil
}

// Insert -
func (m *MongoDBClientStub) Insert(coll mongodb.CollectionID, key []byte, data []byte) error {
	if m.InsertCalled != nil {
		return m.InsertCalled(coll, key, data)
	}

	return nil
}

// PutIfVersion -
func (m *MongoDBClientStub) PutIfVersion(coll mongodb.CollectionID, key []byte, data []byte, expectedVersion int64) error {
	if m.PutIfVersionCalled != nil {
		return m.PutIfVersionCalled(coll, key, data, expectedVersion)
	}

	return nil
}

// Has -
func (m *MongoDBClientStub) Has(coll mongodb.CollectionID, key []byte) error {
	if m.HasCalled != nil {
		return m.HasCalled(coll, key)
	}

	return nil
}

// Remove -
func (m *MongoDBClientStub) Remove(coll mongodb.CollectionID, key []byte) error {
	if m.RemoveCalled != nil {
		return m.RemoveCalled(coll, key)
	}

	return nil
}

// GetIndex -
func (m *MongoDBClientStub) GetIndex(coll mongodb.CollectionID, key []byte) (uint32, error) {
	if m.GetIndexCalled != nil {
		return m.GetIndexCalled(coll, key)
	}

	return 0, nil
}

// IncrementIndex -
func (m *MongoDBClientStub) IncrementIndex(coll mongodb.CollectionID, key []byte) (uint32, error) {
	if m.IncrementIndexCalled != nil {
		return m.IncrementIndexCalled(coll, key)
	}

	return 0, nil
}

// PutIndexIfNotExists -
func (m *MongoDBClientStub) PutIndexIfNotExists(collID mongodb.CollectionID, key []byte, index uint32) error {
	if m.PutIndexIfNotExistsCalled != nil {
		return m.PutIndexIfNotExistsCalled(collID, key, index)
	}

	return nil
}

// GetAllCollectionsIDs -
func (m *MongoDBClientStub) GetAllCollectionsIDs() []mongodb.CollectionID {
	if m.GetAllCollectionsIDsCalled != nil {
		return m.GetAllCollectionsIDsCalled()
	}

	return make([]mongodb.CollectionID, 0)
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
