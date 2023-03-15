package sync

import "sync"

type KeyRWMutex struct {
	mut            sync.RWMutex
	managedMutexes map[string]*RwMutex
}

// NewKeyRWMutex returns a new instance of KeyRWMutex
func NewKeyRWMutex() *KeyRWMutex {
	return &KeyRWMutex{
		managedMutexes: make(map[string]*RwMutex),
	}
}

// RLock locks for read the Mutex for the given key
func (csa *KeyRWMutex) RLock(key string) {
	csa.getMutex(key).RLock()
}

// RUnlock unlocks for read the Mutex for the given key
func (csa *KeyRWMutex) RUnlock(key string) {
	mutex := csa.getMutex(key)
	mutex.RUnlock()

	csa.cleanupMutex(key, mutex)
}

// Lock locks the Mutex for the given key
func (csa *KeyRWMutex) Lock(key string) {
	csa.getMutex(key).Lock()
}

// Unlock unlocks the Mutex for the given key
func (csa *KeyRWMutex) Unlock(key string) {
	mutex := csa.getMutex(key)
	mutex.Unlock()

	csa.cleanupMutex(key, mutex)
}

// getMutex returns the Mutex for the given key
func (csa *KeyRWMutex) getMutex(key string) *RwMutex {
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
func (csa *KeyRWMutex) cleanupMutex(key string, mutex *RwMutex) {
	csa.mut.Lock()
	defer csa.mut.Unlock()
	if mutex.NumLocks() == 0 {
		delete(csa.managedMutexes, key)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (csa *KeyRWMutex) IsInterfaceNil() bool {
	return csa == nil
}
