package testscommon

import (
	"sync"
	"time"

	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
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

// CheckAllowedAndIncreaseTrials -
func (r *RateLimiterMock) CheckAllowedAndIncreaseTrials(key string, _ redis.Mode) (*redis.RateLimiterResult, error) {
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

// SetSecurityModeNoExpire -
func (r *RateLimiterMock) SetSecurityModeNoExpire(key string) error {
	return nil
}

// UnsetSecurityModeNoExpire -
func (r *RateLimiterMock) UnsetSecurityModeNoExpire(key string) error {
	return nil
}

// Reset -
func (r *RateLimiterMock) Reset(key string) error {
	r.mutTrials.Lock()
	defer r.mutTrials.Unlock()

	delete(r.trials, key)

	return nil
}

// DecrementSecurityFailedTrials -
func (r *RateLimiterMock) DecrementSecurityFailedTrials(key string) error {
	r.mutTrials.Lock()
	defer r.mutTrials.Unlock()

	t, exists := r.trials[key]
	if !exists {
		return nil
	}

	r.trials[key] = t - 1

	return nil
}

// Period -
func (r *RateLimiterMock) Period(_ redis.Mode) time.Duration {
	return r.periodLimit
}

// Rate -
func (r *RateLimiterMock) Rate(_ redis.Mode) int {
	return r.maxFailures
}

// ExtendSecurityMode -
func (r *RateLimiterMock) ExtendSecurityMode(_ string) error {
	return nil
}

// IsInterfaceNil -
func (r *RateLimiterMock) IsInterfaceNil() bool {
	return r == nil
}
