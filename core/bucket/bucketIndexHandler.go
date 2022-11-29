package bucket

import (
	"encoding/binary"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
)

const (
	lastIndexKey = "lastAllocatedIndex"
	uint32Bytes  = 4
)

type bucketIndexHandler struct {
	bucket core.Storer
	mut    sync.RWMutex
}

// NewBucketIndexHandler returns a new instance of a bucket index handler
func NewBucketIndexHandler(bucket core.Storer) (*bucketIndexHandler, error) {
	if check.IfNil(bucket) {
		return nil, core.ErrNilBucket
	}

	handler := &bucketIndexHandler{
		bucket: bucket,
	}

	err := bucket.Has([]byte(lastIndexKey))
	if err == nil {
		return handler, nil
	}

	err = handler.saveNewIndex(0)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

// UpdateIndexReturningNext updates the index and returns the new value
func (handler *bucketIndexHandler) UpdateIndexReturningNext() (uint32, error) {
	handler.mut.Lock()
	defer handler.mut.Unlock()

	index, err := handler.getIndex()
	if err != nil {
		return 0, err
	}

	index++

	return index, handler.saveNewIndex(index)
}

// Put adds data to the bucket
func (handler *bucketIndexHandler) Put(key, data []byte) error {
	handler.mut.Lock()
	defer handler.mut.Unlock()

	return handler.bucket.Put(key, data)
}

// Get returns the value for the key from the bucket
func (handler *bucketIndexHandler) Get(key []byte) ([]byte, error) {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	return handler.bucket.Get(key)
}

// Has returns true if the key exists in the bucket
func (handler *bucketIndexHandler) Has(key []byte) error {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	return handler.bucket.Has(key)
}

// Close closes the internal bucket
func (handler *bucketIndexHandler) Close() error {
	handler.mut.Lock()
	defer handler.mut.Unlock()

	return handler.bucket.Close()
}

// must be called under mutex protection
func (handler *bucketIndexHandler) getIndex() (uint32, error) {
	lastIndexBytes, err := handler.bucket.Get([]byte(lastIndexKey))
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(lastIndexBytes), nil
}

// must be called under mutex protection
func (handler *bucketIndexHandler) saveNewIndex(newIndex uint32) error {
	latestIndexBytes := make([]byte, uint32Bytes)
	binary.BigEndian.PutUint32(latestIndexBytes, newIndex)
	return handler.bucket.Put([]byte(lastIndexKey), latestIndexBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *bucketIndexHandler) IsInterfaceNil() bool {
	return handler == nil
}
