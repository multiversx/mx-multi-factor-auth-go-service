package testsCommon

// IndexHandlerStub -
type IndexHandlerStub struct {
	GetIndexCalled func() uint32
}

// GetIndex -
func (stub *IndexHandlerStub) GetIndex() uint32 {
	if stub.GetIndexCalled != nil {
		return stub.GetIndexCalled()
	}
	return 0
}

// IsInterfaceNil -
func (stub *IndexHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
