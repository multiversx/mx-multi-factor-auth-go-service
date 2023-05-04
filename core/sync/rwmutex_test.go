package sync

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRWMutex(t *testing.T) {
	t.Parallel()

	cs := NewRWMutex()
	require.NotNil(t, cs)
	require.Equal(t, uint32(0), cs.cntLocks)
	require.Equal(t, uint32(0), cs.cntRLocks)
}

func TestRWMutex_Lock_Unlock_IsLocked_NumLocks(t *testing.T) {
	t.Parallel()

	cs := &rwMutex{}
	cs.lock()
	cs.updateCounterLock()
	require.Equal(t, uint32(1), cs.numLocks())

	cs.unlock()
	cs.updateCounterUnlock()
	require.Equal(t, uint32(0), cs.numLocks())

	cs.rLock()
	cs.updateCounterRLock()
	require.Equal(t, uint32(1), cs.numLocks())

	cs.rUnlock()
	cs.updateCounterRUnlock()
	require.Equal(t, uint32(0), cs.numLocks())
}

func TestRWMutex_MultipleLocksUnlocks(t *testing.T) {
	t.Parallel()

	cs := &rwMutex{}
	numConcurrentCalls := 500
	wg := sync.WaitGroup{}

	f := func(wg *sync.WaitGroup, cs *rwMutex) {
		cs.updateCounterLock()
		cs.lock()
		_ = cs.numLocks()

		cs.updateCounterUnlock()
		cs.unlock()
		_ = cs.numLocks()

		cs.updateCounterRLock()
		cs.rLock()
		_ = cs.numLocks()

		cs.updateCounterRUnlock()
		cs.rUnlock()
		_ = cs.numLocks()

		wg.Done()
	}

	wg.Add(numConcurrentCalls)

	for i := 1; i <= numConcurrentCalls; i++ {
		go f(&wg, cs)
	}
	wg.Wait()

	require.Equal(t, uint32(0), cs.numLocks())
}
