package sync

import (
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
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
	cs.Lock()
	require.True(t, cs.IsLocked())
	require.Equal(t, uint32(1), cs.NumLocks())

	cs.Unlock()
	require.False(t, cs.IsLocked())
	require.Equal(t, uint32(0), cs.NumLocks())
}

func TestRWMutex_MultipleLocksUnlocks(t *testing.T) {
	t.Parallel()

	cs := &rwMutex{}
	numConcurrentCalls := 100
	wg := sync.WaitGroup{}

	f := func(wg *sync.WaitGroup, cs RWMutexHandler) {
		cs.Lock()
		<-time.After(time.Millisecond * 10)
		cs.Unlock()

		cs.RLock()
		<-time.After(time.Millisecond * 10)
		cs.RUnlock()

		wg.Done()
	}

	wg.Add(numConcurrentCalls)

	for i := 1; i <= numConcurrentCalls; i++ {
		go f(&wg, cs)
		// checking for concurrency issues also with IsLocked and NumLocks
		_ = cs.IsLocked()
		_ = cs.NumLocks()
	}
	wg.Wait()

	require.False(t, cs.IsLocked())
	require.Equal(t, uint32(0), cs.NumLocks())
}

func TestRWMutex_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	cs := &rwMutex{}
	require.False(t, check.IfNil(cs))

	cs = nil
	require.True(t, check.IfNil(cs))

	var cs2 RWMutexHandler
	require.True(t, check.IfNil(cs2))
}
