package bucket

import (
	"encoding/binary"
	"sync"

	"github.com/multiversx/multi-factor-auth-go-service/core"
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

	err = saveNewIndex(handler.bucket, 0)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

// AllocateBucketIndex allocates a new index and returns it
func (handler *bucketIndexHandler) AllocateBucketIndex() (uint32, error) {
	handler.mut.Lock()
	defer handler.mut.Unlock()

	index, err := getIndex(handler.bucket)
	if err != nil {
		return 0, err
	}

	index++

	return index, saveNewIndex(handler.bucket, index)
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

func (handler *bucketIndexHandler) UpdateWithCheck(key []byte, fn func(data interface{}) (interface{}, error)) error {
	data, err := handler.bucket.Get(key)
	if err != nil {
		return nil
	}

	newData, err := fn(data)
	if err != nil {
		return err
	}
	newDataBytes, ok := newData.([]byte)
	if !ok {
		return core.ErrInvalidValue
	}

	return handler.bucket.Put(key, newDataBytes)
}

// GetLastIndex returns the last index that was allocated
func (handler *bucketIndexHandler) GetLastIndex() (uint32, error) {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	return getIndex(handler.bucket)
}

// Close closes the internal bucket
func (handler *bucketIndexHandler) Close() error {
	handler.mut.Lock()
	defer handler.mut.Unlock()

	return handler.bucket.Close()
}

// must be called under mutex protection
func getIndex(storer core.Storer) (uint32, error) {
	lastIndexBytes, err := storer.Get([]byte(lastIndexKey))
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(lastIndexBytes), nil
}

// must be called under mutex protection
func saveNewIndex(storer core.Storer, newIndex uint32) error {
	latestIndexBytes := make([]byte, uint32Bytes)
	binary.BigEndian.PutUint32(latestIndexBytes, newIndex)
	return storer.Put([]byte(lastIndexKey), latestIndexBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *bucketIndexHandler) IsInterfaceNil() bool {
	return handler == nil
}
