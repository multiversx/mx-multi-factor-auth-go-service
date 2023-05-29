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
func (r *RedisMutexMock) Lock() {
	r.mut.Lock()
}

// LockContext -
func (r *RedisMutexMock) LockContext(ctx context.Context) {
	r.mut.Lock()
}

// Unlock -
func (r *RedisMutexMock) Unlock() {
	r.mut.Unlock()
}

// UnlockContext -
func (r *RedisMutexMock) UnlockContext(ctx context.Context) {
	r.mut.Unlock()
}
