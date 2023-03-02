package testscommon

import "time"

// RedisClientWrapperStub -
type RedisClientWrapperStub struct {
	SetCalled    func(key string, value interface{}, expiration time.Duration) (string, error)
	GetCalled    func(key string) (string, error)
	ExistsCalled func(keys string) (int64, error)
	IncrByCalled func(key []byte, increment int64) (int64, error)
	DelCalled    func(keys string) (int64, error)
	CloseCalled  func() error
}

// Set -
func (r *RedisClientWrapperStub) Set(key string, value interface{}, expiration time.Duration) (string, error) {
	if r.SetCalled != nil {
		return r.SetCalled(key, value, expiration)
	}

	return "", nil
}

// Get -
func (r *RedisClientWrapperStub) Get(key string) (string, error) {
	if r.GetCalled != nil {
		return r.GetCalled(key)
	}

	return "", nil
}

// Exists -
func (r *RedisClientWrapperStub) Exists(keys string) (int64, error) {
	if r.ExistsCalled != nil {
		return r.ExistsCalled(keys)
	}

	return 0, nil
}

// IncrBy -
func (r *RedisClientWrapperStub) IncrBy(key []byte, increment int64) (int64, error) {
	if r.IncrByCalled != nil {
		return r.IncrByCalled(key, increment)
	}

	return 0, nil
}

// Del -
func (r *RedisClientWrapperStub) Del(keys string) (int64, error) {
	if r.DelCalled != nil {
		return r.DelCalled(keys)
	}

	return 0, nil
}

// Close -
func (r *RedisClientWrapperStub) Close() error {
	if r.CloseCalled != nil {
		return r.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (r *RedisClientWrapperStub) IsInterfaceNil() bool {
	return r == nil
}
