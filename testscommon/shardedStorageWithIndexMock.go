package testscommon

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/storage"
)

type versionedEntry struct {
	data    []byte
	version int64
}

type shardedStorageWithIndexMock struct {
	mut   sync.RWMutex
	cache map[string]*versionedEntry
}

// NewShardedStorageWithIndexMock -
func NewShardedStorageWithIndexMock() *shardedStorageWithIndexMock {
	return &shardedStorageWithIndexMock{
		cache: make(map[string]*versionedEntry),
	}
}

// AllocateIndex -
func (mock *shardedStorageWithIndexMock) AllocateIndex(_ []byte) (uint32, error) {
	return 0, nil
}

// Put keeps the stored version, like a mongodb $set
func (mock *shardedStorageWithIndexMock) Put(key, data []byte) error {
	mock.mut.Lock()
	defer mock.mut.Unlock()
	entry, ok := mock.cache[string(key)]
	if !ok {
		mock.cache[string(key)] = &versionedEntry{data: data}
		return nil
	}
	entry.data = data
	return nil
}

// Get -
func (mock *shardedStorageWithIndexMock) Get(key []byte) ([]byte, error) {
	data, _, err := mock.GetWithVersion(key)
	return data, err
}

// GetWithVersion -
func (mock *shardedStorageWithIndexMock) GetWithVersion(key []byte) ([]byte, int64, error) {
	mock.mut.RLock()
	defer mock.mut.RUnlock()
	entry, ok := mock.cache[string(key)]
	if !ok {
		return nil, 0, common.ErrKeyNotFound
	}
	return entry.data, entry.version, nil
}

// Insert -
func (mock *shardedStorageWithIndexMock) Insert(key, data []byte) error {
	mock.mut.Lock()
	defer mock.mut.Unlock()
	_, ok := mock.cache[string(key)]
	if ok {
		return storage.ErrKeyAlreadyExists
	}
	mock.cache[string(key)] = &versionedEntry{data: data, version: 1}
	return nil
}

// PutIfVersion -
func (mock *shardedStorageWithIndexMock) PutIfVersion(key, data []byte, expectedVersion int64) error {
	mock.mut.Lock()
	defer mock.mut.Unlock()
	entry, ok := mock.cache[string(key)]
	if !ok || entry.version != expectedVersion {
		return storage.ErrVersionMismatch
	}
	entry.data = data
	entry.version++
	return nil
}

// Has -
func (mock *shardedStorageWithIndexMock) Has(key []byte) error {
	mock.mut.RLock()
	_, ok := mock.cache[string(key)]
	mock.mut.RUnlock()
	if !ok {
		return fmt.Errorf("key not found")
	}
	return nil
}

// Close -
func (mock *shardedStorageWithIndexMock) Close() error {
	return nil
}

// Count -
func (mock *shardedStorageWithIndexMock) Count() (uint32, error) {
	mock.mut.RLock()
	defer mock.mut.RUnlock()
	return uint32(len(mock.cache)), nil
}

// IsInterfaceNil -
func (mock *shardedStorageWithIndexMock) IsInterfaceNil() bool {
	return mock == nil
}
