package sync

import (
	"sync"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// keyRWMutex is a mutex that can be used to lock/unlock a resource identified by a key
type keyRWMutex struct {
	mut            sync.RWMutex
	managedMutexes map[string]*rwMutex
	lockHandler    redis.Locker
}

// NewKeyRWMutex returns a new instance of keyRWMutex
func NewKeyRWMutex(lockHandler redis.Locker) (*keyRWMutex, error) {
	if check.IfNil(lockHandler) {
		return nil, core.ErrNilLocker
	}

	return &keyRWMutex{
		managedMutexes: make(map[string]*rwMutex),
		lockHandler:    lockHandler,
	}, nil
}

// RLock locks for read the Mutex for the given key
func (csa *keyRWMutex) RLock(key string) {
	csa.getForRLock(key).rLock()
}

// RUnlock unlocks for read the Mutex for the given key
func (csa *keyRWMutex) RUnlock(key string) {
	csa.getForRUnlock(key).rUnlock()
	csa.cleanupMutex(key)
}

// Lock locks the Mutex for the given key
func (csa *keyRWMutex) Lock(key string) {
	csa.getForLock(key).lock()
}

// Unlock unlocks the Mutex for the given key
func (csa *keyRWMutex) Unlock(key string) {
	csa.getForUnlock(key).unlock()
	csa.cleanupMutex(key)
}

// getForLock returns the Mutex for the given key, updating the Lock counter
func (csa *keyRWMutex) getForLock(key string) *rwMutex {
	csa.mut.Lock()
	defer csa.mut.Unlock()

	mutex, ok := csa.managedMutexes[key]
	if !ok {
		mutex = csa.newInternalMutex(key)
	}
	mutex.updateCounterLock()

	return mutex
}

// getForRLock returns the Mutex for the given key, updating the RLock counter
func (csa *keyRWMutex) getForRLock(key string) *rwMutex {
	csa.mut.Lock()
	defer csa.mut.Unlock()

	mutex, ok := csa.managedMutexes[key]
	if !ok {
		mutex = csa.newInternalMutex(key)
	}
	mutex.updateCounterRLock()

	return mutex
}

// getForUnlock returns the Mutex for the given key, updating the Unlock counter
func (csa *keyRWMutex) getForUnlock(key string) *rwMutex {
	csa.mut.Lock()
	defer csa.mut.Unlock()

	mutex, ok := csa.managedMutexes[key]
	if ok {
		mutex.updateCounterUnlock()
	}

	return mutex
}

// getForRUnlock returns the Mutex for the given key, updating the RUnlock counter
func (csa *keyRWMutex) getForRUnlock(key string) *rwMutex {
	csa.mut.Lock()
	defer csa.mut.Unlock()

	mutex, ok := csa.managedMutexes[key]
	if ok {
		mutex.updateCounterRUnlock()
	}

	return mutex
}

// newInternalMutex creates a new mutex for the given key and adds it to the map
func (csa *keyRWMutex) newInternalMutex(key string) *rwMutex {
	mutex, ok := csa.managedMutexes[key]
	if !ok {
		m := csa.lockHandler.NewMutex(key)
		mutex = newRWMutex(m)
		csa.managedMutexes[key] = mutex
	}
	return mutex
}

// cleanupMutex removes the mutex from the map if it is not used anymore
func (csa *keyRWMutex) cleanupMutex(key string) {
	csa.mut.Lock()
	defer csa.mut.Unlock()

	mut, ok := csa.managedMutexes[key]
	if ok && mut.numLocks() == 0 {
		delete(csa.managedMutexes, key)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (csa *keyRWMutex) IsInterfaceNil() bool {
	return csa == nil
}
