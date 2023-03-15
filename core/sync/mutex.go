package sync

import "sync"

type rwMutex struct {
	internalMut sync.RWMutex
	cntLocks    uint32
	cntRLocks   uint32

	controlMut sync.RWMutex
}

// Lock locks the rwMutex
func (rm *rwMutex) Lock() {
	rm.internalMut.Lock()
	rm.cntLocks++
	rm.internalMut.Unlock()

	rm.controlMut.Lock()
}

// Unlock unlocks the rwMutex
func (rm *rwMutex) Unlock() {
	rm.internalMut.Lock()
	rm.cntLocks--
	rm.internalMut.Unlock()

	rm.controlMut.Unlock()
}

func (rm *rwMutex) RLock() {
	rm.internalMut.Lock()
	rm.cntRLocks++
	rm.internalMut.Unlock()

	rm.controlMut.RLock()
}

func (rm *rwMutex) RUnlock() {
	rm.internalMut.Lock()
	rm.cntRLocks--
	rm.internalMut.Unlock()

	rm.controlMut.RUnlock()
}

// IsLocked returns true if the rwMutex is locked
func (rm *rwMutex) IsLocked() bool {
	rm.internalMut.RLock()
	cntLock := rm.cntLocks
	cntRLock := rm.cntRLocks
	rm.internalMut.RUnlock()

	return cntLock > 0 || cntRLock > 0
}

// NumLocks returns the number of locks on the rwMutex
func (rm *rwMutex) NumLocks() uint32 {
	rm.internalMut.RLock()
	cntLocks := rm.cntLocks
	cntRLocks := rm.cntRLocks
	rm.internalMut.RUnlock()

	return cntLocks + cntRLocks
}

// IsInterfaceNil returns true if there is no value under the interface
func (rm *rwMutex) IsInterfaceNil() bool {
	return rm == nil
}
