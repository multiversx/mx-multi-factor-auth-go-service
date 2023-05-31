package sync

type baseMutex struct {
	cntLocks  int32
	cntRLocks int32
}

// newBaseMutex returns a new instance of baseMutex
func newBaseMutex() *baseMutex {
	return &baseMutex{}
}

func (bm *baseMutex) updateCounterLock() {
	bm.cntLocks++
}

func (bm *baseMutex) updateCounterRLock() {
	bm.cntRLocks++
}

func (bm *baseMutex) updateCounterUnlock() {
	bm.cntLocks--
}

func (bm *baseMutex) updateCounterRUnlock() {
	bm.cntRLocks--
}

// numLocks returns the number of locks on the rwMutex
func (bm *baseMutex) numLocks() int32 {
	cntLocks := bm.cntLocks
	cntRLocks := bm.cntRLocks

	return cntLocks + cntRLocks
}
