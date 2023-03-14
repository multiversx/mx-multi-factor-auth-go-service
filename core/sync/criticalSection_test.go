package sync

import (
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestCriticalSection_Lock_Unlock_IsLocked_NumLocks(t *testing.T) {
	cs := &criticalSection{}
	cs.Lock()
	require.True(t, cs.IsLocked())
	require.Equal(t, uint32(1), cs.NumLocks())
	cs.Unlock()
	require.False(t, cs.IsLocked())
	require.Equal(t, uint32(0), cs.NumLocks())
}

func TestCriticalSection_MultipleLocksUnlocks(t *testing.T) {
	cs := &criticalSection{}
	numConcurrentCalls := 100
	wg := sync.WaitGroup{}

	f := func(wg *sync.WaitGroup, cs CriticalSection) {
		cs.Lock()
		<-time.After(time.Millisecond * 10)
		cs.Unlock()

		wg.Done()
	}

	wg.Add(numConcurrentCalls)

	for i := 1; i <= numConcurrentCalls; i++ {
		go f(&wg, cs)
		// checking for concurrency issues
		_ = cs.IsLocked()
		_ = cs.NumLocks()
	}
	wg.Wait()

	require.False(t, cs.IsLocked())
	require.Equal(t, uint32(0), cs.NumLocks())
}

func TestCriticalSection_IsInterfaceNil(t *testing.T) {
	cs := &criticalSection{}
	require.False(t, check.IfNil(cs))

	cs = nil
	require.True(t, check.IfNil(cs))

	var cs2 CriticalSection
	require.True(t, check.IfNil(cs2))
}
