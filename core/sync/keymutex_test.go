package sync

import (
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestNewKeyMutex(t *testing.T) {
	csa := NewKeyMutex()
	require.NotNil(t, csa)
}

func TestKeyMutex_Lock_Unlock(t *testing.T) {
	csa := NewKeyMutex()
	require.NotNil(t, csa)
	require.Len(t, csa.(*keyMutex).managedMutexes, 0)
	csa.Lock("id1")
	require.Len(t, csa.(*keyMutex).managedMutexes, 1)
	csa.Lock("id2")
	require.Len(t, csa.(*keyMutex).managedMutexes, 2)
	csa.Unlock("id1")
	require.Len(t, csa.(*keyMutex).managedMutexes, 1)
	csa.Unlock("id2")
	require.Len(t, csa.(*keyMutex).managedMutexes, 0)
}

func TestKeyMutex_RLock_RUnlock(t *testing.T) {
	csa := NewKeyMutex()
	require.NotNil(t, csa)
	require.Len(t, csa.(*keyMutex).managedMutexes, 0)
	csa.RLock("id1")
	require.Len(t, csa.(*keyMutex).managedMutexes, 1)
	csa.RLock("id2")
	require.Len(t, csa.(*keyMutex).managedMutexes, 2)
	csa.RUnlock("id1")
	require.Len(t, csa.(*keyMutex).managedMutexes, 1)
	csa.RUnlock("id2")
	require.Len(t, csa.(*keyMutex).managedMutexes, 0)
}

func TestKeyMutex_IsInterfaceNil(t *testing.T) {
	csa := NewKeyMutex()
	require.False(t, check.IfNil(csa))

	csa = nil
	require.True(t, check.IfNil(csa))

	var csa2 KeyMutex
	require.True(t, check.IfNil(csa2))
}

func TestKeyMutex_ConcurrencyMultipleCriticalSections(t *testing.T) {
	wg := sync.WaitGroup{}
	csa := NewKeyMutex()
	require.NotNil(t, csa)

	f := func(wg *sync.WaitGroup, id string) {
		csa.Lock(id)
		<-time.After(time.Millisecond * 10)
		csa.Unlock(id)

		csa.RLock(id)
		<-time.After(time.Millisecond * 10)
		csa.RUnlock(id)

		wg.Done()
	}

	numConcurrentCalls := 100
	ids := []string{"id1", "id2", "id3", "id4", "id5", "id6", "id7", "id8", "id9", "id10"}
	wg.Add(numConcurrentCalls)

	for i := 1; i <= numConcurrentCalls; i++ {
		go f(&wg, ids[i%len(ids)])
	}
	wg.Wait()

	require.Len(t, csa.(*keyMutex).managedMutexes, 0)
}
