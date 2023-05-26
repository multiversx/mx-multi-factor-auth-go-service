package testscommon

import (
	"context"
	"sync"

	"github.com/multiversx/multi-factor-auth-go-service/redis"
)

// RedLockSyncerMock implements Locker interface
type RedLockSyncerMock struct{}

// NewMutex -
func (r *RedLockSyncerMock) NewMutex(name string) redis.RedLockMutex {
	return NewRedLockMutexMock()
}

// IsInterfaceNil -
func (r *RedLockSyncerMock) IsInterfaceNil() bool {
	return r == nil
}

// RedLockMutexMock -
type RedLockMutexMock struct {
	mut sync.RWMutex
}

// NewRedLockMutexMock -
func NewRedLockMutexMock() *RedLockMutexMock {
	return &RedLockMutexMock{}
}

// Lock -
func (r *RedLockMutexMock) Lock() error {
	r.mut.Lock()
	return nil
}

// LockContext -
func (r *RedLockMutexMock) LockContext(ctx context.Context) error {
	return nil
}

// Unlock -
func (r *RedLockMutexMock) Unlock() (bool, error) {
	r.mut.Unlock()
	return true, nil
}

// UnlockContext -
func (r *RedLockMutexMock) UnlockContext(ctx context.Context) (bool, error) {
	return true, nil
}
