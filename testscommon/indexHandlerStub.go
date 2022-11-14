package testscommon

// IndexHandlerStub -
type IndexHandlerStub struct {
	AllocateIndexCalled func() (uint32, error)
}

// AllocateIndex -
func (stub *IndexHandlerStub) AllocateIndex() (uint32, error) {
	if stub.AllocateIndexCalled != nil {
		return stub.AllocateIndexCalled()
	}
	return 0, nil
}

// IsInterfaceNil -
func (stub *IndexHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
