package testscommon

import (
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
)

// MongoDBClientStub implemented mongodb client wraper interface
type MongoDBClientStub struct {
	PutCalled                      func(coll mongodb.CollectionID, key []byte, data []byte) error
	GetCalled                      func(coll mongodb.CollectionID, key []byte) ([]byte, error)
	HasCalled                      func(coll mongodb.CollectionID, key []byte) error
	RemoveCalled                   func(coll mongodb.CollectionID, key []byte) error
	IncrementWithTransactionCalled func(coll mongodb.CollectionID, key []byte) (uint32, error)
	CloseCalled                    func() error
	ReadWriteWithCheckCalled       func(
		collID mongodb.CollectionID,
		key []byte,
		checker func(data interface{}) (interface{}, error),
	) error
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

// IncrementWithTransaction -
func (m *MongoDBClientStub) IncrementWithTransaction(coll mongodb.CollectionID, key []byte) (uint32, error) {
	if m.IncrementWithTransactionCalled != nil {
		return m.IncrementWithTransactionCalled(coll, key)
	}

	return 0, nil
}

// ReadWriteWithCheck -
func (m *MongoDBClientStub) ReadWriteWithCheck(
	collID mongodb.CollectionID,
	key []byte,
	checker func(data interface{}) (interface{}, error),
) error {
	if m.ReadWriteWithCheckCalled != nil {
		return m.ReadWriteWithCheckCalled(collID, key, checker)
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
