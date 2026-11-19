package testscommon

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/mx-multi-factor-auth-go-service/mongodb"
)

type collCache struct {
	cache map[string]*versionedEntry
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
			cache: make(map[string]*versionedEntry),
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

	entry, ok := collection.cache[string(key)]
	if !ok {
		collection.cache[string(key)] = &versionedEntry{data: data}
		return nil
	}
	entry.data = data

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

	entry, ok := collection.cache[string(key)]
	if !ok {
		return nil, common.ErrKeyNotFound
	}

	return entry.data, nil
}

// GetWithVersion -
func (m *mongoDBClientMock) GetWithVersion(coll mongodb.CollectionID, key []byte) ([]byte, int64, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	collection, ok := m.collections[coll]
	if !ok {
		return nil, 0, mongodb.ErrCollectionNotFound
	}

	entry, ok := collection.cache[string(key)]
	if !ok {
		return nil, 0, common.ErrKeyNotFound
	}

	return entry.data, entry.version, nil
}

// Insert -
func (m *mongoDBClientMock) Insert(coll mongodb.CollectionID, key []byte, data []byte) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	collection, ok := m.collections[coll]
	if !ok {
		return mongodb.ErrCollectionNotFound
	}

	_, ok = collection.cache[string(key)]
	if ok {
		return storage.ErrKeyAlreadyExists
	}
	collection.cache[string(key)] = &versionedEntry{data: data, version: 1}

	return nil
}

// PutIfVersion -
func (m *mongoDBClientMock) PutIfVersion(coll mongodb.CollectionID, key []byte, data []byte, expectedVersion int64) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	collection, ok := m.collections[coll]
	if !ok {
		return mongodb.ErrCollectionNotFound
	}

	entry, ok := collection.cache[string(key)]
	if !ok || entry.version != expectedVersion {
		return storage.ErrVersionMismatch
	}
	entry.data = data
	entry.version++

	return nil
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
