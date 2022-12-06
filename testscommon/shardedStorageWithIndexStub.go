package testscommon

// ShardedStorageWithIndexStub -
type ShardedStorageWithIndexStub struct {
	AllocateIndexCalled            func(address []byte) (uint32, error)
	PutCalled                      func(key, data []byte) error
	GetCalled                      func(key []byte) ([]byte, error)
	HasCalled                      func(key []byte) error
	CloseCalled                    func() error
	UpdateIndexReturningNextCalled func(address []byte) (uint32, error)
	NumberOfBucketsCalled          func() uint32
}

// AllocateIndex -
func (stub *ShardedStorageWithIndexStub) AllocateIndex(address []byte) (uint32, error) {
	if stub.AllocateIndexCalled != nil {
		return stub.AllocateIndexCalled(address)
	}
	return 0, nil
}

// Put -
func (stub *ShardedStorageWithIndexStub) Put(key, data []byte) error {
	if stub.PutCalled != nil {
		return stub.PutCalled(key, data)
	}
	return nil
}

// Get -
func (stub *ShardedStorageWithIndexStub) Get(key []byte) ([]byte, error) {
	if stub.GetCalled != nil {
		return stub.GetCalled(key)
	}
	return nil, nil
}

// Has -
func (stub *ShardedStorageWithIndexStub) Has(key []byte) error {
	if stub.HasCalled != nil {
		return stub.HasCalled(key)
	}
	return nil
}

// Close -
func (stub *ShardedStorageWithIndexStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// UpdateIndexReturningNext -
func (stub *ShardedStorageWithIndexStub) UpdateIndexReturningNext(address []byte) (uint32, error) {
	if stub.UpdateIndexReturningNextCalled != nil {
		return stub.UpdateIndexReturningNextCalled(address)
	}
	return 0, nil
}

// NumberOfBuckets -
func (stub *ShardedStorageWithIndexStub) NumberOfBuckets() uint32 {
	if stub.NumberOfBucketsCalled != nil {
		return stub.NumberOfBucketsCalled()
	}
	return 0
}

// IsInterfaceNil -
func (stub *ShardedStorageWithIndexStub) IsInterfaceNil() bool {
	return stub == nil
}
