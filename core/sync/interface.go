package sync

// RWMutexHandler is the interface that defines the methods that can be used on a mutex
type RWMutexHandler interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
	IsLocked() bool
	NumLocks() uint32
	IsInterfaceNil() bool
}

// KeyRWMutexHandler is a mutex that can be used to lock/unlock a resource identified by a key
type KeyRWMutexHandler interface {
	Lock(key string)
	Unlock(key string)
	RLock(key string)
	RUnlock(key string)
	IsInterfaceNil() bool
}
