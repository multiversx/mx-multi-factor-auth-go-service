package testscommon

// SignatureVerifierStub -
type SignatureVerifierStub struct {
	VerifyCalled func(pk []byte, msg []byte, skBytes []byte) error
}

// Verify -
func (stub *SignatureVerifierStub) Verify(pk []byte, msg []byte, skBytes []byte) error {
	if stub.VerifyCalled != nil {
		return stub.VerifyCalled(pk, msg, skBytes)
	}
	return nil
}

// IsInterfaceNil -
func (stub *SignatureVerifierStub) IsInterfaceNil() bool {
	return stub == nil
}
