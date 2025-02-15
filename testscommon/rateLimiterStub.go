package testscommon

import (
	"time"

	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
)

// RateLimiterStub -
type RateLimiterStub struct {
	CheckAllowedAndIncreaseTrialsCalled func(key string, mode redis.Mode) (*redis.RateLimiterResult, error)
	DecrementSecurityFailuresCalled     func(key string) error
	ResetCalled                         func(key string) error
	PeriodCalled                        func(mode redis.Mode) time.Duration
	RateCalled                          func(mode redis.Mode) int
	SetSecurityModeNoExpireCalled       func(key string) error
	UnsetSecurityModeNoExpireCalled     func(key string) error
	ExtendSecurityModeCalled            func(key string) error
}

// CheckAllowedAndIncreaseTrials -
func (r *RateLimiterStub) CheckAllowedAndIncreaseTrials(key string, mode redis.Mode) (*redis.RateLimiterResult, error) {
	if r.CheckAllowedAndIncreaseTrialsCalled != nil {
		return r.CheckAllowedAndIncreaseTrialsCalled(key, mode)
	}

	return nil, nil
}

// DecrementSecurityFailedTrials -
func (r *RateLimiterStub) DecrementSecurityFailedTrials(key string) error {
	if r.DecrementSecurityFailuresCalled != nil {
		return r.DecrementSecurityFailuresCalled(key)
	}

	return nil
}

// SetSecurityModeNoExpire -
func (r *RateLimiterStub) SetSecurityModeNoExpire(key string) error {
	if r.SetSecurityModeNoExpireCalled != nil {
		return r.SetSecurityModeNoExpireCalled(key)
	}
	return nil
}

// UnsetSecurityModeNoExpire -
func (r *RateLimiterStub) UnsetSecurityModeNoExpire(key string) error {
	if r.UnsetSecurityModeNoExpireCalled != nil {
		return r.UnsetSecurityModeNoExpireCalled(key)
	}
	return nil
}

// Reset -
func (r *RateLimiterStub) Reset(key string) error {
	if r.ResetCalled != nil {
		return r.ResetCalled(key)
	}

	return nil
}

// Period -
func (r *RateLimiterStub) Period(mode redis.Mode) time.Duration {
	if r.PeriodCalled != nil {
		return r.PeriodCalled(mode)
	}

	return 0
}

// Rate -
func (r *RateLimiterStub) Rate(mode redis.Mode) int {
	if r.RateCalled != nil {
		return r.RateCalled(mode)
	}

	return 0
}

// ExtendSecurityMode -
func (r *RateLimiterStub) ExtendSecurityMode(key string) error {
	if r.ExtendSecurityModeCalled != nil {
		return r.ExtendSecurityModeCalled(key)
	}

	return nil
}

// IsInterfaceNil -
func (r *RateLimiterStub) IsInterfaceNil() bool {
	return r == nil
}
