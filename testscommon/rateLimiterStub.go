package testscommon

import "time"

// RateLimiterStub -
type RateLimiterStub struct {
	CheckAllowedCalled func(key string, maxFailures int, maxDuration time.Duration) (int, error)
	ResetCalled        func(key string) error
}

// CheckAllowed -
func (r *RateLimiterStub) CheckAllowed(key string, maxFailures int, maxDuration time.Duration) (int, error) {
	if r.CheckAllowedCalled != nil {
		return r.CheckAllowedCalled(key, maxFailures, maxDuration)
	}

	return 0, nil
}

// Reset -
func (r *RateLimiterStub) Reset(key string) error {
	if r.ResetCalled != nil {
		return r.ResetCalled(key)
	}

	return nil
}

// IsInterfaceNil -
func (r *RateLimiterStub) IsInterfaceNil() bool {
	return r == nil
}
