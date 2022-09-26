package testsCommon

import "github.com/ElrondNetwork/elrond-sdk-erdgo/data"

// GuardianStub -
type GuardianStub struct {
	ValidateAndSendCalled func(transaction data.Transaction) (string, error)
	GetAddressCalled      func() string
	AddUserCalled         func(address string)
	HasUserCalled         func(address string) bool
	RemoveUserCalled      func(address string)
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

// AddUser -
func (g *GuardianStub) AddUser(address string) {
	if g.AddUserCalled != nil {
		g.AddUserCalled(address)
	}
}

// HasUser -
func (g *GuardianStub) HasUser(address string) bool {
	if g.HasUserCalled != nil {
		return g.HasUserCalled(address)
	}

	return false
}

// RemoveUser -
func (g *GuardianStub) RemoveUser(address string) {
	if g.RemoveUserCalled != nil {
		g.RemoveUserCalled(address)
	}
}

// IsInterfaceNil -
func (g *GuardianStub) IsInterfaceNil() bool {
	return g == nil
}
