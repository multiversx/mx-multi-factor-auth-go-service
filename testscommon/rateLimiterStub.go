package testscommon

import (
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/redis"
)

// RateLimiterStub -
type RateLimiterStub struct {
	CheckAllowedCalled func(key string) (*redis.RateLimiterResult, error)
	ResetCalled        func(key string) error
	PeriodCalled       func() time.Duration
	RateCalled         func() int
}

// CheckAllowed -
func (r *RateLimiterStub) CheckAllowed(key string) (*redis.RateLimiterResult, error) {
	if r.CheckAllowedCalled != nil {
		return r.CheckAllowedCalled(key)
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

// IsInterfaceNil -
func (r *RateLimiterStub) IsInterfaceNil() bool {
	return r == nil
}
