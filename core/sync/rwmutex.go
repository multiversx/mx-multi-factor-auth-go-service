package sync

import "sync"

// rwMutex is a mutex that can be used to lock/unlock a resource
type rwMutex struct {
	internalMut sync.RWMutex
	cntLocks    uint32
	cntRLocks   uint32

	controlMut sync.RWMutex
}

// NewRWMutex returns a new instance of rwMutex
func NewRWMutex() *rwMutex {
	return &rwMutex{}
}

func (rm *rwMutex) updateCounterLock() {
	rm.internalMut.Lock()
	rm.cntLocks++
	rm.internalMut.Unlock()
}

func (rm *rwMutex) updateCounterRLock() {
	rm.internalMut.Lock()
	rm.cntRLocks++
	rm.internalMut.Unlock()
}

func (rm *rwMutex) updateCounterUnlock() {
	rm.internalMut.Lock()
	rm.cntLocks--
	rm.internalMut.Unlock()
}

func (rm *rwMutex) updateCounterRUnlock() {
	rm.internalMut.Lock()
	rm.cntRLocks--
	rm.internalMut.Unlock()
}

// Lock locks the rwMutex
func (rm *rwMutex) lock() {
	rm.controlMut.Lock()
}

// Unlock unlocks the rwMutex
func (rm *rwMutex) unlock() {
	rm.controlMut.Unlock()
}

// RLock locks for read the rwMutex
func (rm *rwMutex) rLock() {
	rm.controlMut.RLock()
}

// RUnlock unlocks for read the rwMutex
func (rm *rwMutex) rUnlock() {
	rm.controlMut.RUnlock()
}

// NumLocks returns the number of locks on the rwMutex
func (rm *rwMutex) numLocks() uint32 {
	rm.internalMut.RLock()
	cntLocks := rm.cntLocks
	cntRLocks := rm.cntRLocks
	rm.internalMut.RUnlock()

	return cntLocks + cntRLocks
}
