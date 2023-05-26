package redis

import (
	"time"

	"github.com/go-redsync/redsync/v4"
)

const defaultLockTimeoutInSec = 3600

type redSyncWrapper struct {
	redSyncer *redsync.Redsync
}

// NewRedSyncWrapper will create a new relock syncer component
func NewRedSyncWrapper(redSyncer *redsync.Redsync) *redSyncWrapper {
	return &redSyncWrapper{redSyncer: redSyncer}
}

// NewMutex will create a new mutex
func (r *redSyncWrapper) NewMutex(name string) RedLockMutex {
	opt := redsync.WithExpiry(time.Duration(defaultLockTimeoutInSec) * time.Second)
	return r.redSyncer.NewMutex(name, opt)
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *redSyncWrapper) IsInterfaceNil() bool {
	return r == nil
}
