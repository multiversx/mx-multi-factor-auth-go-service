package testscommon

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/multiversx/mx-multi-factor-auth-go-service/mongodb"
)

type collCache struct {
	cache map[string][]byte
}

type mongoDBClientMock struct {
	mut            sync.RWMutex
	collections    map[mongodb.CollectionID]*collCache
	collectionsIDs []mongodb.CollectionID
}

// NewMongoDBClientMock -
func NewMongoDBClientMock(numCollections uint32) *mongoDBClientMock {
	collections := make(map[mongodb.CollectionID]*collCache)
	collectionIDs := make([]mongodb.CollectionID, 0)

	for i := uint32(0); i < numCollections; i++ {
		collID := mongodb.CollectionID(fmt.Sprintf("%s_%d", string(mongodb.UsersCollectionID), i))
		collections[collID] = &collCache{
			cache: make(map[string][]byte),
		}
		collectionIDs = append(collectionIDs, collID)
	}

	return &mongoDBClientMock{
		collections:    collections,
		collectionsIDs: collectionIDs,
	}
}

// Put -
func (m *mongoDBClientMock) Put(coll mongodb.CollectionID, key []byte, data []byte) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	collection, ok := m.collections[coll]
	if !ok {
		return mongodb.ErrCollectionNotFound
	}

	collection.cache[string(key)] = data

	return nil
}

// Get -
func (m *mongoDBClientMock) Get(coll mongodb.CollectionID, key []byte) ([]byte, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	collection, ok := m.collections[coll]
	if !ok {
		return nil, mongodb.ErrCollectionNotFound
	}

	data, ok := collection.cache[string(key)]
	if !ok {
		return nil, common.ErrKeyNotFound
	}

	return data, nil
}

// Has -
func (m *mongoDBClientMock) Has(coll mongodb.CollectionID, key []byte) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	collection, ok := m.collections[coll]
	if !ok {
		return mongodb.ErrCollectionNotFound
	}

	_, ok = collection.cache[string(key)]
	if !ok {
		return common.ErrKeyNotFound
	}

	return nil
}

// Remove -
func (m *mongoDBClientMock) Remove(coll mongodb.CollectionID, key []byte) error {
	return nil
}

// GetIndex -
func (m *mongoDBClientMock) GetIndex(coll mongodb.CollectionID, key []byte) (uint32, error) {
	return 0, nil
}

// IncrementIndex -
func (m *mongoDBClientMock) IncrementIndex(coll mongodb.CollectionID, key []byte) (uint32, error) {
	return 0, nil
}

// PutIndexIfNotExists -
func (m *mongoDBClientMock) PutIndexIfNotExists(collID mongodb.CollectionID, key []byte, index uint32) error {
	return nil
}

// GetAllCollectionsIDs -
func (m *mongoDBClientMock) GetAllCollectionsIDs() []mongodb.CollectionID {
	return m.collectionsIDs
}

// Close -
func (m *mongoDBClientMock) Close() error {
	return nil
}

// IsInterfaceNil -
func (m *mongoDBClientMock) IsInterfaceNil() bool {
	return m == nil
}
