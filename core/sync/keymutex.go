package sync

import "sync"

// keyRWMutex is a mutex that can be used to lock/unlock a resource identified by a key
type keyRWMutex struct {
	mut            sync.RWMutex
	managedMutexes map[string]*rwMutex
}

// NewKeyRWMutex returns a new instance of keyRWMutex
func NewKeyRWMutex() *keyRWMutex {
	return &keyRWMutex{
		managedMutexes: make(map[string]*rwMutex),
	}
}

// RLock locks for read the Mutex for the given key
func (csa *keyRWMutex) RLock(key string) {
	csa.getMutex(key).RLock()
}

// RUnlock unlocks for read the Mutex for the given key
func (csa *keyRWMutex) RUnlock(key string) {
	mutex := csa.getMutex(key)
	mutex.RUnlock()

	csa.cleanupMutex(key, mutex)
}

// Lock locks the Mutex for the given key
func (csa *keyRWMutex) Lock(key string) {
	csa.getMutex(key).Lock()
}

// Unlock unlocks the Mutex for the given key
func (csa *keyRWMutex) Unlock(key string) {
	mutex := csa.getMutex(key)
	mutex.Unlock()

	csa.cleanupMutex(key, mutex)
}

// getMutex returns the Mutex for the given key
func (csa *keyRWMutex) getMutex(key string) *rwMutex {
	csa.mut.RLock()
	mutex, ok := csa.managedMutexes[key]
	csa.mut.RUnlock()
	if ok {
		return mutex
	}

	csa.mut.Lock()
	mutex, ok = csa.managedMutexes[key]
	if !ok {
		mutex = NewRWMutex()
		csa.managedMutexes[key] = mutex
	}
	csa.mut.Unlock()

	return mutex
}

// cleanupMutex removes the mutex from the map if it is not used anymore
func (csa *keyRWMutex) cleanupMutex(key string, mutex *rwMutex) {
	csa.mut.Lock()
	defer csa.mut.Unlock()
	if mutex.NumLocks() == 0 {
		delete(csa.managedMutexes, key)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (csa *keyRWMutex) IsInterfaceNil() bool {
	return csa == nil
}
