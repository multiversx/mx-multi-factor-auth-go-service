package sync

import (
	"sync"

	"github.com/multiversx/multi-factor-auth-go-service/redis"
)

// rwMutex is a mutex that can be used to lock/unlock a resource
// this component is not concurrent safe, concurrent accesses need to be managed by the caller
type rwMutex struct {
	*baseMutex
	controlMut sync.RWMutex
}

// newRWMutex returns a new instance of rwMutex
func newRWMutex(m redis.Mutex) *rwMutex {
	return &rwMutex{
		baseMutex: newBaseMutex(),
	}
}

// lock locks the rwMutex
func (rm *rwMutex) lock() error {
	rm.controlMut.Lock()
	return nil
}

// unlock unlocks the rwMutex
func (rm *rwMutex) unlock() error {
	rm.controlMut.Unlock()
	return nil
}

// rLock locks for read the rwMutex
func (rm *rwMutex) rLock() error {
	rm.controlMut.RLock()
	return nil
}

// rUnlock unlocks for read the rwMutex
func (rm *rwMutex) rUnlock() error {
	rm.controlMut.RUnlock()
	return nil
}
