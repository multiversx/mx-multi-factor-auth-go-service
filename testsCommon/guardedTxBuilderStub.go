package testsCommon

import "github.com/ElrondNetwork/elrond-sdk-erdgo/data"

// GuardedTxBuilderStub -
type GuardedTxBuilderStub struct {
	ApplyGuardianSignatureCalled func(skGuardianBytes []byte, tx *data.Transaction) error
}

// ApplyGuardianSignature -
func (stub *GuardedTxBuilderStub) ApplyGuardianSignature(skGuardianBytes []byte, tx *data.Transaction) error {
	if stub.ApplyGuardianSignatureCalled != nil {
		return stub.ApplyGuardianSignatureCalled(skGuardianBytes, tx)
	}
	return nil
}

// IsInterfaceNil -
func (stub *GuardedTxBuilderStub) IsInterfaceNil() bool {
	return stub == nil
}
