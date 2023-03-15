package sync

import "sync"

type keyMutex struct {
	mut            sync.RWMutex
	managedMutexes map[string]*rwMutex
}

// NewKeyMutex returns a new instance of CriticalmutexsAggregator
func NewKeyMutex() KeyMutex {
	return &keyMutex{
		managedMutexes: make(map[string]*rwMutex),
	}
}

func (csa *keyMutex) RLock(key string) {
	csa.getMutex(key).RLock()
}

func (csa *keyMutex) RUnlock(key string) {
	mutex := csa.getMutex(key)
	mutex.RUnlock()

	csa.cleanupMutex(key, mutex)
}

func (csa *keyMutex) Lock(key string) {
	csa.getMutex(key).Lock()
}

func (csa *keyMutex) Unlock(key string) {
	mutex := csa.getMutex(key)
	mutex.Unlock()

	csa.cleanupMutex(key, mutex)
}

// getMutex returns the critical rwMutex for the given key
func (csa *keyMutex) getMutex(key string) Mutex {
	csa.mut.RLock()
	mutex, ok := csa.managedMutexes[key]
	csa.mut.RUnlock()
	if ok {
		return mutex
	}

	csa.mut.Lock()
	mutex, ok = csa.managedMutexes[key]
	if !ok {
		mutex = &rwMutex{}
		csa.managedMutexes[key] = mutex
	}
	csa.mut.Unlock()

	return mutex
}

func (csa *keyMutex) cleanupMutex(key string, mutex Mutex) {
	csa.mut.Lock()
	defer csa.mut.Unlock()
	if mutex.NumLocks() == 0 {
		delete(csa.managedMutexes, key)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (csa *keyMutex) IsInterfaceNil() bool {
	return csa == nil
}
