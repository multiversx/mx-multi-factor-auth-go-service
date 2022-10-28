package testsCommon

// IndexHandlerStub -
type IndexHandlerStub struct {
	AllocateIndexCalled func() uint32
	RevertIndexCalled   func()
}

// AllocateIndex -
func (stub *IndexHandlerStub) AllocateIndex() uint32 {
	if stub.AllocateIndexCalled != nil {
		return stub.AllocateIndexCalled()
	}
	return 0
}

// RevertIndex -
func (stub *IndexHandlerStub) RevertIndex() {
	if stub.RevertIndexCalled != nil {
		stub.RevertIndexCalled()
	}
}

// IsInterfaceNil -
func (stub *IndexHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
