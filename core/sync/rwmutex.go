package sync

import (
	"github.com/multiversx/multi-factor-auth-go-service/redis"
)

// rwMutex is a mutex that can be used to lock/unlock a resource
// this component is not concurrent safe, concurrent accesses need to be managed by the caller
type rwMutex struct {
	cntLocks  int32
	cntRLocks int32

	controlMut redis.RedLockMutex
}

// newRWMutex returns a new instance of rwMutex
func newRWMutex(m redis.RedLockMutex) *rwMutex {
	return &rwMutex{controlMut: m}
}

func (rm *rwMutex) updateCounterLock() {
	rm.cntLocks++
}

func (rm *rwMutex) updateCounterRLock() {
	rm.cntRLocks++
}

func (rm *rwMutex) updateCounterUnlock() {
	rm.cntLocks--
}

func (rm *rwMutex) updateCounterRUnlock() {
	rm.cntRLocks--
}

// lock locks the rwMutex
func (rm *rwMutex) lock() {
	rm.controlMut.Lock()
}

// unlock unlocks the rwMutex
func (rm *rwMutex) unlock() {
	rm.controlMut.Unlock()
}

// rLock locks for read the rwMutex
func (rm *rwMutex) rLock() {
	rm.controlMut.Lock()
}

// rUnlock unlocks for read the rwMutex
func (rm *rwMutex) rUnlock() {
	rm.controlMut.Unlock()
}

// numLocks returns the number of locks on the rwMutex
func (rm *rwMutex) numLocks() int32 {
	cntLocks := rm.cntLocks
	cntRLocks := rm.cntRLocks

	return cntLocks + cntRLocks
}
