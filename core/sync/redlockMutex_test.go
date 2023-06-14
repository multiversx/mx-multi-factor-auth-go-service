package sync

import (
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewRedlockMutex(t *testing.T) {
	t.Parallel()

	rm := newRedLockMutex(&testscommon.RedisMutexMock{})
	require.NotNil(t, rm)
	require.Equal(t, int32(0), rm.cntLocks)
	require.Equal(t, int32(0), rm.cntRLocks)
}

func TestRedLockMutex_Opearations(t *testing.T) {
	t.Parallel()

	rm := newRedLockMutex(&testscommon.RedisMutexMock{})
	rm.lock()
	rm.updateCounterLock()
	require.Equal(t, int32(1), rm.numLocks())

	rm.unlock()
	rm.updateCounterUnlock()
	require.Equal(t, int32(0), rm.numLocks())

	rm.rLock()
	rm.updateCounterRLock()
	require.Equal(t, int32(1), rm.numLocks())

	rm.rUnlock()
	rm.updateCounterRUnlock()
	require.Equal(t, int32(0), rm.numLocks())
}
