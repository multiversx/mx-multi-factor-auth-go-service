package sync

import "github.com/multiversx/multi-factor-auth-go-service/redis"

type redlockMutex struct {
	*baseMutex
	controlMut redis.Mutex
}

// newRedLockMutex returns a new instance of redlock mutex
func newRedLockMutex(m redis.Mutex) *redlockMutex {
	return &redlockMutex{
		baseMutex:  newBaseMutex(),
		controlMut: m,
	}
}

// lock locks the rwMutex
func (rm *redlockMutex) lock() error {
	return rm.controlMut.Lock()
}

// unlock unlocks the rwMutex
func (rm *redlockMutex) unlock() error {
	return rm.controlMut.Unlock()
}

// rLock locks for read the rwMutex
func (rm *redlockMutex) rLock() error {
	return rm.controlMut.Lock()
}

// rUnlock unlocks for read the rwMutex
func (rm *redlockMutex) rUnlock() error {
	return rm.controlMut.Unlock()
}
