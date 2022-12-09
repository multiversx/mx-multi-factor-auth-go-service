package testscommon

type BucketIDProviderStub struct {
	GetBucketForAddressCalled func(address []byte) uint32
}

// GetBucketForAddress -
func (stub *BucketIDProviderStub) GetBucketForAddress(address []byte) uint32 {
	if stub.GetBucketForAddressCalled != nil {
		return stub.GetBucketForAddressCalled(address)
	}
	return 0
}

// IsInterfaceNil -
func (stub *BucketIDProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
