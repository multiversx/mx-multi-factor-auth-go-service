package testscommon

import (
	"context"
	"sync"

	"github.com/multiversx/multi-factor-auth-go-service/redis"
)

// RedLockSyncerMock implements Locker interface
type RedLockSyncerMock struct{}

// NewMutex -
func (r *RedLockSyncerMock) NewMutex(name string) redis.Mutex {
	return NewRedisMutexMock()
}

// IsInterfaceNil -
func (r *RedLockSyncerMock) IsInterfaceNil() bool {
	return r == nil
}

// RedisMutexMock -
type RedisMutexMock struct {
	mut sync.RWMutex
}

// NewRedisMutexMock -
func NewRedisMutexMock() *RedisMutexMock {
	return &RedisMutexMock{}
}

// Lock -
func (r *RedisMutexMock) Lock() error {
	r.mut.Lock()
	return nil
}

// Unlock -
func (r *RedisMutexMock) Unlock() error {
	r.mut.Unlock()
	return nil
}

// RedLockMutexStub -
type RedLockMutexStub struct {
	UnlockContextCalled func(ctx context.Context) (bool, error)
	LockContextCalled   func(ctx context.Context) error
}

// LockContext -
func (r *RedLockMutexStub) LockContext(ctx context.Context) error {
	if r.LockContextCalled != nil {
		return r.LockContextCalled(ctx)
	}

	return nil
}

// UnlockContext -
func (r *RedLockMutexStub) UnlockContext(ctx context.Context) (bool, error) {
	if r.UnlockContextCalled != nil {
		return r.UnlockContextCalled(ctx)
	}

	return false, nil
}
