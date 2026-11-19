package testscommon

import (
	"sync"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
)

// InterleavingStorage wraps a storage and runs a hook once, right after the first versioned read
type InterleavingStorage struct {
	core.StorageWithIndex
	once           sync.Once
	afterFirstRead func()
}

// NewInterleavingStorage -
func NewInterleavingStorage(storage core.StorageWithIndex, afterFirstRead func()) *InterleavingStorage {
	return &InterleavingStorage{
		StorageWithIndex: storage,
		afterFirstRead:   afterFirstRead,
	}
}

// GetWithVersion reads first, then runs the hook, so the caller continues with stale data
func (s *InterleavingStorage) GetWithVersion(key []byte) ([]byte, int64, error) {
	data, version, err := s.StorageWithIndex.GetWithVersion(key)
	s.once.Do(func() {
		if s.afterFirstRead != nil {
			s.afterFirstRead()
		}
	})

	return data, version, err
}

// IsInterfaceNil -
func (s *InterleavingStorage) IsInterfaceNil() bool {
	return s == nil
}
