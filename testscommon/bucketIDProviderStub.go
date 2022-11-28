package testscommon

type BucketIDProviderStub struct {
	GetIDFromAddressCalled func(address []byte) uint32
}

// GetIDFromAddress -
func (stub *BucketIDProviderStub) GetIDFromAddress(address []byte) uint32 {
	if stub.GetIDFromAddressCalled != nil {
		return stub.GetIDFromAddressCalled(address)
	}
	return 0
}

// IsInterfaceNil -
func (stub *BucketIDProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
