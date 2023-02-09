package testscommon

import (
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// GuardedTxBuilderStub -
type GuardedTxBuilderStub struct {
	ApplyGuardianSignatureCalled func(cryptoHolderGuardian core.CryptoComponentsHolder, tx *data.Transaction) error
}

// ApplyGuardianSignature -
func (stub *GuardedTxBuilderStub) ApplyGuardianSignature(cryptoHolderGuardian core.CryptoComponentsHolder, tx *data.Transaction) error {
	if stub.ApplyGuardianSignatureCalled != nil {
		return stub.ApplyGuardianSignatureCalled(cryptoHolderGuardian, tx)
	}
	return nil
}

// IsInterfaceNil -
func (stub *GuardedTxBuilderStub) IsInterfaceNil() bool {
	return stub == nil
}
