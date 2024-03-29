package bucket

import (
	"encoding/binary"
	"sync"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
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

// AllocateBucketIndex allocates a new index and returns it
func (handler *bucketIndexHandler) AllocateBucketIndex() (uint32, error) {
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
	return handler.bucket.Put(key, data)
}

// Get returns the value for the key from the bucket
func (handler *bucketIndexHandler) Get(key []byte) ([]byte, error) {
	return handler.bucket.Get(key)
}

// Has returns true if the key exists in the bucket
func (handler *bucketIndexHandler) Has(key []byte) error {
	return handler.bucket.Has(key)
}

// GetLastIndex returns the last index that was allocated
func (handler *bucketIndexHandler) GetLastIndex() (uint32, error) {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	return handler.getIndex()
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
