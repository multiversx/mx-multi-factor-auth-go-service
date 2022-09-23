package testsCommon

import "github.com/ElrondNetwork/elrond-sdk-erdgo/data"

// GuardianStub -
type GuardianStub struct {
	ValidateAndSendCalled func(transaction data.Transaction) (string, error)
	GetAddressCalled      func() string
}

// ValidateAndSend -
func (g *GuardianStub) ValidateAndSend(transaction data.Transaction) (string, error) {
	if g.ValidateAndSendCalled != nil {
		return g.ValidateAndSendCalled(transaction)
	}
	return "", nil
}

// GetAddress -
func (g *GuardianStub) GetAddress() string {
	if g.GetAddressCalled != nil {
		return g.GetAddressCalled()
	}
	return ""
}

// IsInterfaceNil -
func (g *GuardianStub) IsInterfaceNil() bool {
	return g == nil
}
