package sync

type Mutex interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
	IsLocked() bool
	NumLocks() uint32
	IsInterfaceNil() bool
}

type KeyMutex interface {
	Lock(key string)
	Unlock(key string)
	RLock(key string)
	RUnlock(key string)
	IsInterfaceNil() bool
}
