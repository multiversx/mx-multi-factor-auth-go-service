package testsCommon

// StorerStub -
type StorerStub struct {
	PutCalled    func(key, data []byte) error
	GetCalled    func(key []byte) ([]byte, error)
	HasCalled    func(key []byte) bool
	RemoveCalled func(key []byte) error
	LenCalled    func() int
	CloseCalled  func() error
}

// Put -
func (stub *StorerStub) Put(key, data []byte) error {
	if stub.PutCalled != nil {
		return stub.PutCalled(key, data)
	}
	return nil
}

// Get -
func (stub *StorerStub) Get(key []byte) ([]byte, error) {
	if stub.GetCalled != nil {
		return stub.GetCalled(key)
	}
	return make([]byte, 0), nil
}

// Has -
func (stub *StorerStub) Has(key []byte) bool {
	if stub.HasCalled != nil {
		return stub.HasCalled(key)
	}
	return false
}

// Remove -
func (stub *StorerStub) Remove(key []byte) error {
	if stub.RemoveCalled != nil {
		return stub.RemoveCalled(key)
	}
	return nil
}

// Len -
func (stub *StorerStub) Len() int {
	if stub.LenCalled != nil {
		return stub.LenCalled()
	}
	return 0
}

// Close -
func (stub *StorerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *StorerStub) IsInterfaceNil() bool {
	return stub == nil
}
