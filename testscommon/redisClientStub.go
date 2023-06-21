package testscommon

import (
	"context"
	"time"
)

// RedisClientStub -
type RedisClientStub struct {
	SetEntryIfNotExistingCalled   func(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error)
	DeleteCalled                  func(ctx context.Context, key string) error
	DecrementCalled               func(ctx context.Context, key string) (int64, error)
	ExpireTimeCalled              func(ctx context.Context, key string) (time.Duration, error)
	DecrementWithExpireTimeCalled func(ctx context.Context, key string) (int64, time.Duration, error)
}

// SetEntryIfNotExisting -
func (r *RedisClientStub) SetEntryIfNotExisting(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
	if r.SetEntryIfNotExistingCalled != nil {
		return r.SetEntryIfNotExistingCalled(ctx, key, value, ttl)
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

// Decrement -
func (r *RedisClientStub) Decrement(ctx context.Context, key string) (int64, error) {
	if r.DecrementCalled != nil {
		return r.DecrementCalled(ctx, key)
	}

	return 0, nil
}

// ExpireTime -
func (r *RedisClientStub) ExpireTime(ctx context.Context, key string) (time.Duration, error) {
	if r.ExpireTimeCalled != nil {
		return r.ExpireTimeCalled(ctx, key)
	}

	return 0, nil
}

// DecrementWithExpireTime -
func (r *RedisClientStub) DecrementWithExpireTime(ctx context.Context, key string) (int64, time.Duration, error) {
	if r.DecrementWithExpireTimeCalled != nil {
		return r.DecrementWithExpireTimeCalled(ctx, key)
	}

	return 0, 0, nil
}

// IsInterfaceNil -
func (r *RedisClientStub) IsInterfaceNil() bool {
	return r == nil
}
