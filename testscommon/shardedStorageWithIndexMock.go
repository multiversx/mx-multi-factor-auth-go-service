package testscommon

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-storage-go/common"
)

type shardedStorageWithIndexMock struct {
	mut   sync.RWMutex
	cache map[string][]byte
}

// NewShardedStorageWithIndexMock -
func NewShardedStorageWithIndexMock() *shardedStorageWithIndexMock {
	return &shardedStorageWithIndexMock{
		cache: make(map[string][]byte),
	}
}

// AllocateIndex -
func (mock *shardedStorageWithIndexMock) AllocateIndex(_ []byte) (uint32, error) {
	return 0, nil
}

// Put -
func (mock *shardedStorageWithIndexMock) Put(key, data []byte) error {
	mock.mut.Lock()
	mock.cache[string(key)] = data
	mock.mut.Unlock()
	return nil
}

// Get -
func (mock *shardedStorageWithIndexMock) Get(key []byte) ([]byte, error) {
	mock.mut.RLock()
	data, ok := mock.cache[string(key)]
	mock.mut.RUnlock()
	if !ok {
		return nil, common.ErrKeyNotFound
	}
	return data, nil
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

func (mock *shardedStorageWithIndexMock) UpdateWithCheck(key []byte, fn func(data interface{}) (interface{}, error)) error {
	data, err := mock.Get(key)
	if err != nil {
		return err
	}

	newData, err := fn(data)
	if err != nil {
		return err
	}

	return mock.Put(key, newData.([]byte))
}

// IsInterfaceNil -
func (mock *shardedStorageWithIndexMock) IsInterfaceNil() bool {
	return mock == nil
}
