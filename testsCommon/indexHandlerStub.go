package testsCommon

// IndexHandlerStub -
type IndexHandlerStub struct {
	AllocateIndexCalled func() uint32
}

// AllocateIndex -
func (stub *IndexHandlerStub) AllocateIndex() uint32 {
	if stub.AllocateIndexCalled != nil {
		return stub.AllocateIndexCalled()
	}
	return 0
}

// IsInterfaceNil -
func (stub *IndexHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
