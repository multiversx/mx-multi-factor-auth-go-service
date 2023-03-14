package sync

type CriticalSection interface {
	Lock()
	Unlock()
	IsLocked() bool
	NumLocks() uint32
	IsInterfaceNil() bool
}

type CriticalSectionsAggregator interface {
	Lock(key string)
	Unlock(key string)
	IsInterfaceNil() bool
}
