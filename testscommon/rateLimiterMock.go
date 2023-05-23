package testscommon

import (
	"sync"
	"time"
)

// RateLimiterMock -
type RateLimiterMock struct {
	trials    map[string]int
	mutTrials sync.RWMutex
}

func NewRateLimiterMock() *RateLimiterMock {
	return &RateLimiterMock{
		trials: make(map[string]int),
	}
}

// CheckAllowed -
func (r *RateLimiterMock) CheckAllowed(key string, maxFailures int, maxDuration time.Duration) (int, error) {
	r.mutTrials.Lock()
	defer r.mutTrials.Unlock()

	_, exists := r.trials[key]
	if !exists {
		r.trials[key] = 0
		return maxFailures, nil
	}

	if r.trials[key] < maxFailures {
		r.trials[key]++
	}

	remaining := maxFailures - r.trials[key]

	return remaining, nil
}

// Reset -
func (r *RateLimiterMock) Reset(key string) error {
	r.mutTrials.Lock()
	defer r.mutTrials.Unlock()

	delete(r.trials, key)

	return nil
}

// IsInterfaceNil -
func (r *RateLimiterMock) IsInterfaceNil() bool {
	return r == nil
}
