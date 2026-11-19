package testscommon

// BucketIndexHandlerStub -
type BucketIndexHandlerStub struct {
	PutCalled                 func(key, data []byte) error
	GetCalled                 func(key []byte) ([]byte, error)
	GetWithVersionCalled      func(key []byte) ([]byte, int64, error)
	InsertCalled              func(key, data []byte) error
	PutIfVersionCalled        func(key, data []byte, expectedVersion int64) error
	HasCalled                 func(key []byte) error
	CloseCalled               func() error
	AllocateBucketIndexCalled func() (uint32, error)
	GetLastIndexCalled        func() (uint32, error)
}

// Put -
func (stub *BucketIndexHandlerStub) Put(key, data []byte) error {
	if stub.PutCalled != nil {
		return stub.PutCalled(key, data)
	}
	return nil
}

// Get -
func (stub *BucketIndexHandlerStub) Get(key []byte) ([]byte, error) {
	if stub.GetCalled != nil {
		return stub.GetCalled(key)
	}
	return nil, nil
}

// GetWithVersion -
func (stub *BucketIndexHandlerStub) GetWithVersion(key []byte) ([]byte, int64, error) {
	if stub.GetWithVersionCalled != nil {
		return stub.GetWithVersionCalled(key)
	}
	return nil, 0, nil
}

// Insert -
func (stub *BucketIndexHandlerStub) Insert(key, data []byte) error {
	if stub.InsertCalled != nil {
		return stub.InsertCalled(key, data)
	}
	return nil
}

// PutIfVersion -
func (stub *BucketIndexHandlerStub) PutIfVersion(key, data []byte, expectedVersion int64) error {
	if stub.PutIfVersionCalled != nil {
		return stub.PutIfVersionCalled(key, data, expectedVersion)
	}
	return nil
}

// Has -
func (stub *BucketIndexHandlerStub) Has(key []byte) error {
	if stub.HasCalled != nil {
		return stub.HasCalled(key)
	}
	return nil
}

// Close -
func (stub *BucketIndexHandlerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// AllocateBucketIndex -
func (stub *BucketIndexHandlerStub) AllocateBucketIndex() (uint32, error) {
	if stub.AllocateBucketIndexCalled != nil {
		return stub.AllocateBucketIndexCalled()
	}
	return 0, nil
}

// GetLastIndex -
func (stub *BucketIndexHandlerStub) GetLastIndex() (uint32, error) {
	if stub.GetLastIndexCalled != nil {
		return stub.GetLastIndexCalled()
	}
	return 0, nil
}

// IsInterfaceNil -
func (stub *BucketIndexHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
