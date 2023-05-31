package sync

// KeyRWMutexHandler is a mutex that can be used to lock/unlock a resource identified by a key
type KeyRWMutexHandler interface {
	Lock(key string)
	Unlock(key string)
	RLock(key string)
	RUnlock(key string)
	IsInterfaceNil() bool
}

type keyMutex interface {
	updateCounterLock()
	updateCounterRLock()
	updateCounterUnlock()
	updateCounterRUnlock()
	numLocks() int32
	lock() error
	unlock() error
	rLock() error
	rUnlock() error
}
