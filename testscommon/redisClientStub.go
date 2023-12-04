package testscommon

import (
	"context"
	"time"
)

// RedisClientStub -
type RedisClientStub struct {
	IncrementCalled              func(ctx context.Context, key string) (int64, error)
	SetExpireIfNotExistingCalled func(ctx context.Context, key string, ttl time.Duration) (bool, error)
	DeleteCalled                 func(ctx context.Context, key string) error
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

// SetExpireIfNotExisting -
func (r *RedisClientStub) SetExpireIfNotExisting(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if r.SetExpireIfNotExistingCalled != nil {
		return r.SetExpireIfNotExistingCalled(ctx, key, ttl)
	}

	return false, nil
}

// Delete -
func (r *RedisClientStub) Delete(ctx context.Context, key string) error {
	if r.DeleteCalled != nil {
		return r.DeleteCalled(ctx, key)
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
