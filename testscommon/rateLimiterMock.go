package testscommon

import (
	"sync"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/redis"
)

// RateLimiterMock -
type RateLimiterMock struct {
	trials      map[string]int
	mutTrials   sync.RWMutex
	maxFailures int
	periodLimit time.Duration
}

// NewRateLimiterMock -
func NewRateLimiterMock(maxFailures int, periodLimit int) *RateLimiterMock {
	return &RateLimiterMock{
		trials:      make(map[string]int),
		maxFailures: maxFailures,
		periodLimit: time.Duration(periodLimit) * time.Second,
	}
}

// CheckAllowedAndDecreaseTrials -
func (r *RateLimiterMock) CheckAllowedAndDecreaseTrials(key string) (*redis.RateLimiterResult, error) {
	r.mutTrials.Lock()
	defer r.mutTrials.Unlock()

	_, exists := r.trials[key]
	if !exists {
		r.trials[key] = 0
		return &redis.RateLimiterResult{Allowed: true, Remaining: r.maxFailures}, nil
	}

	if r.trials[key] < r.maxFailures {
		r.trials[key]++
	}

	remaining := r.maxFailures - r.trials[key]

	allowed := true
	if remaining == 0 {
		allowed = false
	}

	return &redis.RateLimiterResult{Allowed: allowed, Remaining: remaining}, nil
}

// Reset -
func (r *RateLimiterMock) Reset(key string) error {
	r.mutTrials.Lock()
	defer r.mutTrials.Unlock()

	delete(r.trials, key)

	return nil
}

// Period -
func (r *RateLimiterMock) Period() time.Duration {
	return r.periodLimit
}

// Rate -
func (r *RateLimiterMock) Rate() int {
	return r.maxFailures
}

// IsInterfaceNil -
func (r *RateLimiterMock) IsInterfaceNil() bool {
	return r == nil
}
