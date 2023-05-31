package redis

import (
	"context"
	"time"
)

type redlockMutexWrapper struct {
	mutex            RedLockMutex
	operationTimeout time.Duration
}

// NewRedLockMutexWrapper will create a new instance of redlock mutex wrapper component
func NewRedLockMutexWrapper(mutex RedLockMutex, operationTimeout time.Duration) *redlockMutexWrapper {
	return &redlockMutexWrapper{
		mutex:            mutex,
		operationTimeout: operationTimeout,
	}
}

// Lock will try to lock distributed redis mutex
func (rmw *redlockMutexWrapper) Lock() error {
	ctx, cancel := context.WithTimeout(context.Background(), rmw.operationTimeout)
	defer cancel()

	err := rmw.mutex.LockContext(ctx)
	if err != nil {
		log.Error("failed to lock mutex with context", "error", err.Error())
		return err
	}

	return nil
}

// Unlock will try to unlock redis mutex
func (rmw *redlockMutexWrapper) Unlock() error {
	ctx, cancel := context.WithTimeout(context.Background(), rmw.operationTimeout)
	defer cancel()

	ok, err := rmw.mutex.UnlockContext(ctx)
	if err != nil {
		log.Error("failed to unlock mutex with context", "error", err.Error())
		return err
	}

	if !ok {
		log.Warn("did not manage to unlock mutex: no mutex to unlock")
	}

	return nil
}
