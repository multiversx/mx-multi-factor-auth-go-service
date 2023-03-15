package sync

import "sync"

// RwMutex is a mutex that can be used to lock/unlock a resource
type RwMutex struct {
	internalMut sync.RWMutex
	cntLocks    uint32
	cntRLocks   uint32

	controlMut sync.RWMutex
}

// NewRWMutex returns a new instance of RwMutex
func NewRWMutex() *RwMutex {
	return &RwMutex{}
}

// Lock locks the RwMutex
func (rm *RwMutex) Lock() {
	rm.internalMut.Lock()
	rm.cntLocks++
	rm.internalMut.Unlock()

	rm.controlMut.Lock()
}

// Unlock unlocks the RwMutex
func (rm *RwMutex) Unlock() {
	rm.internalMut.Lock()
	rm.cntLocks--
	rm.internalMut.Unlock()

	rm.controlMut.Unlock()
}

func (rm *RwMutex) RLock() {
	rm.internalMut.Lock()
	rm.cntRLocks++
	rm.internalMut.Unlock()

	rm.controlMut.RLock()
}

func (rm *RwMutex) RUnlock() {
	rm.internalMut.Lock()
	rm.cntRLocks--
	rm.internalMut.Unlock()

	rm.controlMut.RUnlock()
}

// IsLocked returns true if the RwMutex is locked
func (rm *RwMutex) IsLocked() bool {
	rm.internalMut.RLock()
	cntLock := rm.cntLocks
	cntRLock := rm.cntRLocks
	rm.internalMut.RUnlock()

	return cntLock > 0 || cntRLock > 0
}

// NumLocks returns the number of locks on the RwMutex
func (rm *RwMutex) NumLocks() uint32 {
	rm.internalMut.RLock()
	cntLocks := rm.cntLocks
	cntRLocks := rm.cntRLocks
	rm.internalMut.RUnlock()

	return cntLocks + cntRLocks
}

// IsInterfaceNil returns true if there is no value under the interface
func (rm *RwMutex) IsInterfaceNil() bool {
	return rm == nil
}
