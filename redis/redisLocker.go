package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/multiversx/multi-factor-auth-go-service/core"
)

const minLockTimeExpiryInSec = 1

var errNilRedSyncer = errors.New("nil red syncer")

type lockerWrapper struct {
	redSyncer           *redsync.Redsync
	lockTimeExpiryInSec time.Duration
}

// NewRedisLockerWrapper will create a new redis locker wrapper component
func NewRedisLockerWrapper(redSyncer *redsync.Redsync, lockTimeExpiry uint64) (*lockerWrapper, error) {
	if redSyncer == nil {
		return nil, errNilRedSyncer
	}
	if lockTimeExpiry < minLockTimeExpiryInSec {
		return nil, fmt.Errorf("%w for LockTimeExpiryInSec, received %d, min expected %d", core.ErrInvalidValue, lockTimeExpiry, minLockTimeExpiryInSec)
	}

	return &lockerWrapper{
		redSyncer:           redSyncer,
		lockTimeExpiryInSec: time.Duration(lockTimeExpiry) * time.Second,
	}, nil
}

// NewMutex will create a new mutex
func (r *lockerWrapper) NewMutex(name string) Mutex {
	opt := redsync.WithExpiry(r.lockTimeExpiryInSec)
	mutex := r.redSyncer.NewMutex(name, opt)
	return newRedLockMutexWrapper(mutex)
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *lockerWrapper) IsInterfaceNil() bool {
	return r == nil
}

type redlockMutexWrapper struct {
	mutex RedLockMutex
}

func newRedLockMutexWrapper(mutex RedLockMutex) *redlockMutexWrapper {
	return &redlockMutexWrapper{
		mutex: mutex,
	}
}

// Lock will try to lock distributed redis mutex
func (rmw *redlockMutexWrapper) Lock() {
	err := rmw.mutex.Lock()
	if err != nil {
		log.Warn("failed to lock mutex", "error", err.Error())
	}
}

// LockContext will try to lock distributed redis mutex
func (rmw *redlockMutexWrapper) LockContext(ctx context.Context) {
	err := rmw.mutex.LockContext(ctx)
	if err != nil {
		log.Warn("failed to lock mutex with context", "error", err.Error())
	}
}

// Unlock will try to unlock redis mutex
func (rmw *redlockMutexWrapper) Unlock() {
	_, err := rmw.mutex.Unlock()
	if err != nil {
		log.Warn("failed to unlock mutex", "error", err.Error())
	}
}

// UnlockContext will try to unlock redis mutex
func (rmw *redlockMutexWrapper) UnlockContext(ctx context.Context) {
	_, err := rmw.mutex.UnlockContext(ctx)
	if err != nil {
		log.Warn("failed to unlock mutex with context", "error", err.Error())
	}
}
