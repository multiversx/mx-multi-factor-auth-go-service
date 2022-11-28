package testscommon

// IndexHandlerStub -
type IndexHandlerStub struct {
	AllocateIndexCalled func(address []byte) (uint32, error)
}

// AllocateIndex -
func (stub *IndexHandlerStub) AllocateIndex(address []byte) (uint32, error) {
	if stub.AllocateIndexCalled != nil {
		return stub.AllocateIndexCalled(address)
	}
	return 0, nil
}

// IsInterfaceNil -
func (stub *IndexHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
