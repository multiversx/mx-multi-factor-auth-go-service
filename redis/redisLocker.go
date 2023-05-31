package redis

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/multiversx/multi-factor-auth-go-service/core"
)

const minLockTimeExpiryInSec = 1

// ErrNilRedSyncer signals that a nil red syncer has been provided
var ErrNilRedSyncer = errors.New("nil red syncer")

type ArgsRedisLockerWrapper struct {
	RedSyncer             *redsync.Redsync
	LockTimeExpiry        uint64
	OperationTimeoutInSec uint64
}

type lockerWrapper struct {
	redSyncer           *redsync.Redsync
	lockTimeExpiryInSec time.Duration
	operationTimeout    time.Duration
}

// NewRedisLockerWrapper will create a new redis locker wrapper component
func NewRedisLockerWrapper(args ArgsRedisLockerWrapper) (*lockerWrapper, error) {
	err := checkRedisLockerArgs(args)
	if err != nil {
		return nil, err
	}

	return &lockerWrapper{
		redSyncer:           args.RedSyncer,
		lockTimeExpiryInSec: time.Duration(args.LockTimeExpiry) * time.Second,
		operationTimeout:    time.Duration(args.OperationTimeoutInSec) * time.Second,
	}, nil
}

func checkRedisLockerArgs(args ArgsRedisLockerWrapper) error {
	if args.RedSyncer == nil {
		return ErrNilRedSyncer
	}
	if args.LockTimeExpiry < minLockTimeExpiryInSec {
		return fmt.Errorf("%w for LockTimeExpiryInSec, received %d, min expected %d", core.ErrInvalidValue, args.LockTimeExpiry, minLockTimeExpiryInSec)
	}
	if args.OperationTimeoutInSec < minOperationTimeoutInSec {
		return fmt.Errorf("%w for OperationTimeoutInSec, received %d, min expected %d", core.ErrInvalidValue, args.OperationTimeoutInSec, minOperationTimeoutInSec)
	}

	return nil
}

// NewMutex will create a new mutex
func (r *lockerWrapper) NewMutex(name string) Mutex {
	opt := redsync.WithExpiry(r.lockTimeExpiryInSec)
	mutex := r.redSyncer.NewMutex(name, opt)
	return NewRedLockMutexWrapper(mutex, r.operationTimeout)
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *lockerWrapper) IsInterfaceNil() bool {
	return r == nil
}
