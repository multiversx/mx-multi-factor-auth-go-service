package sync

import "sync"

type criticalSection struct {
	mut            sync.RWMutex
	mutCritSection sync.RWMutex
	cnt            uint32
}

// Lock locks the criticalSection
func (cs *criticalSection) Lock() {
	cs.mut.Lock()
	cs.cnt++
	cs.mut.Unlock()

	cs.mutCritSection.Lock()
}

// Unlock unlocks the criticalSection
func (cs *criticalSection) Unlock() {
	cs.mut.Lock()
	cs.cnt--
	cs.mut.Unlock()

	cs.mutCritSection.Unlock()
}

// IsLocked returns true if the criticalSection is locked
func (cs *criticalSection) IsLocked() bool {
	cs.mut.RLock()
	cnt := cs.cnt
	cs.mut.RUnlock()

	return cnt > 0
}

// NumLocks returns the number of locks on the criticalSection
func (cs *criticalSection) NumLocks() uint32 {
	cs.mut.RLock()
	cnt := cs.cnt
	cs.mut.RUnlock()

	return cnt
}

// IsInterfaceNil returns true if there is no value under the interface
func (cs *criticalSection) IsInterfaceNil() bool {
	return cs == nil
}
