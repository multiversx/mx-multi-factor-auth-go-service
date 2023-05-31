package redis_test

import (
	"errors"
	"testing"

	"github.com/go-redsync/redsync/v4"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/stretchr/testify/require"
)

func TestNewRedisLockerWrapper(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := redis.ArgsRedisLockerWrapper{
			RedSyncer:             &redsync.Redsync{},
			LockTimeExpiry:        10,
			OperationTimeoutInSec: 10,
		}

		locker, err := redis.NewRedisLockerWrapper(args)
		require.Nil(t, err)
		require.NotNil(t, locker)
		require.False(t, locker.IsInterfaceNil())
		require.NotNil(t, locker.NewMutex("key"))
	})

	t.Run("nil red syncer", func(t *testing.T) {
		t.Parallel()

		args := redis.ArgsRedisLockerWrapper{
			RedSyncer:             nil,
			LockTimeExpiry:        10,
			OperationTimeoutInSec: 10,
		}

		locker, err := redis.NewRedisLockerWrapper(args)
		require.Equal(t, redis.ErrNilRedSyncer, err)
		require.Nil(t, locker)
	})

	t.Run("invalid expiry time", func(t *testing.T) {
		t.Parallel()

		args := redis.ArgsRedisLockerWrapper{
			RedSyncer:             &redsync.Redsync{},
			LockTimeExpiry:        0,
			OperationTimeoutInSec: 10,
		}

		locker, err := redis.NewRedisLockerWrapper(args)
		require.True(t, errors.Is(err, core.ErrInvalidValue))
		require.Nil(t, locker)
	})

	t.Run("invalid operation time", func(t *testing.T) {
		t.Parallel()

		args := redis.ArgsRedisLockerWrapper{
			RedSyncer:             &redsync.Redsync{},
			LockTimeExpiry:        10,
			OperationTimeoutInSec: 0,
		}

		locker, err := redis.NewRedisLockerWrapper(args)
		require.True(t, errors.Is(err, core.ErrInvalidValue))
		require.Nil(t, locker)
	})
}
