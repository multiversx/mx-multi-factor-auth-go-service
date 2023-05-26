package redis

import (
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
func (r *lockerWrapper) NewMutex(name string) RedLockMutex {
	opt := redsync.WithExpiry(r.lockTimeExpiryInSec)
	return r.redSyncer.NewMutex(name, opt)
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *lockerWrapper) IsInterfaceNil() bool {
	return r == nil
}
