package testscommon

import (
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/redis"
)

// RateLimiterStub -
type RateLimiterStub struct {
	CheckAllowedAndIncreaseTrialsCalled func(key string) (*redis.RateLimiterResult, error)
	ResetCalled                         func(key string) error
	PeriodCalled                        func() time.Duration
	RateCalled                          func() int
}

// CheckAllowedAndIncreaseTrials -
func (r *RateLimiterStub) CheckAllowedAndIncreaseTrials(key string) (*redis.RateLimiterResult, error) {
	if r.CheckAllowedAndIncreaseTrialsCalled != nil {
		return r.CheckAllowedAndIncreaseTrialsCalled(key)
	}

	return nil, nil
}

// Reset -
func (r *RateLimiterStub) Reset(key string) error {
	if r.ResetCalled != nil {
		return r.ResetCalled(key)
	}

	return nil
}

// Period -
func (r *RateLimiterStub) Period() time.Duration {
	if r.PeriodCalled != nil {
		return r.PeriodCalled()
	}

	return 0
}

// Rate -
func (r *RateLimiterStub) Rate() int {
	if r.RateCalled != nil {
		return r.RateCalled()
	}

	return 0
}

// IsInterfaceNil -
func (r *RateLimiterStub) IsInterfaceNil() bool {
	return r == nil
}
