package testsCommon

// PubkeyConverterStub -
type PubkeyConverterStub struct {
	LenCalled    func() int
	DecodeCalled func(humanReadable string) ([]byte, error)
	EncodeCalled func(pkBytes []byte) string
}

// Len -
func (pcs *PubkeyConverterStub) Len() int {
	if pcs.LenCalled != nil {
		return pcs.LenCalled()
	}

	return 0
}

// Decode -
func (pcs *PubkeyConverterStub) Decode(humanReadable string) ([]byte, error) {
	if pcs.DecodeCalled != nil {
		return pcs.DecodeCalled(humanReadable)
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil -
func (pcs *PubkeyConverterStub) IsInterfaceNil() bool {
	return pcs == nil
}
