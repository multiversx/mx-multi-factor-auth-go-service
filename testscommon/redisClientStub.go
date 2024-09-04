package testscommon

import (
	"context"
	"time"
)

// RedisClientStub -
type RedisClientStub struct {
	IncrementCalled              func(ctx context.Context, key string) (int64, error)
	DecrementCalled              func(ctx context.Context, key string) (int64, error)
	SetExpireCalled              func(ctx context.Context, key string, ttl time.Duration) (bool, error)
	SetExpireIfNotExistsCalled   func(ctx context.Context, key string, ttl time.Duration) (bool, error)
	SetGreaterExpireTTLCalled    func(ctx context.Context, key string, ttl time.Duration) (bool, error)
	ResetCounterAndKeepTTLCalled func(ctx context.Context, key string) error
	ExpireTimeCalled             func(ctx context.Context, key string) (time.Duration, error)
	IsConnectedCalled            func(ctx context.Context) bool
}

// Increment -
func (r *RedisClientStub) Increment(ctx context.Context, key string) (int64, error) {
	if r.IncrementCalled != nil {
		return r.IncrementCalled(ctx, key)
	}

	return 0, nil
}

// Decrement -
func (r *RedisClientStub) Decrement(ctx context.Context, key string) (int64, error) {
	if r.DecrementCalled != nil {
		return r.DecrementCalled(ctx, key)
	}
	return 0, nil
}

// SetExpire -
func (r *RedisClientStub) SetExpire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if r.SetExpireCalled != nil {
		return r.SetExpireCalled(ctx, key, ttl)
	}

	return false, nil
}

// SetExpireIfNotExists -
func (r *RedisClientStub) SetExpireIfNotExists(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if r.SetExpireIfNotExistsCalled != nil {
		return r.SetExpireIfNotExistsCalled(ctx, key, ttl)
	}

	return false, nil
}

// SetGreaterExpireTTL -
func (r *RedisClientStub) SetGreaterExpireTTL(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if r.SetGreaterExpireTTLCalled != nil {
		return r.SetGreaterExpireTTLCalled(ctx, key, ttl)
	}

	return true, nil
}

// ResetCounterAndKeepTTL -
func (r *RedisClientStub) ResetCounterAndKeepTTL(ctx context.Context, key string) error {
	if r.ResetCounterAndKeepTTLCalled != nil {
		return r.ResetCounterAndKeepTTLCalled(ctx, key)
	}

	return nil
}

// ExpireTime -
func (r *RedisClientStub) ExpireTime(ctx context.Context, key string) (time.Duration, error) {
	if r.ExpireTimeCalled != nil {
		return r.ExpireTimeCalled(ctx, key)
	}

	return 0, nil
}

// IsConnected -
func (r *RedisClientStub) IsConnected(ctx context.Context) bool {
	if r.IsConnectedCalled != nil {
		return r.IsConnectedCalled(ctx)
	}

	return false
}

// IsInterfaceNil -
func (r *RedisClientStub) IsInterfaceNil() bool {
	return r == nil
}
