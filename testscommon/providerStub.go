package testscommon

// ProviderStub -
type ProviderStub struct {
	ValidateCodeCalled func(account, guardian, userCode string) error
	RegisterUserCalled func(accountAddress, accountTag, guardian string) ([]byte, error)
}

// ValidateCode -
func (ps *ProviderStub) ValidateCode(account, guardian, userCode string) error {
	if ps.ValidateCodeCalled != nil {
		return ps.ValidateCodeCalled(account, guardian, userCode)
	}
	return nil
}

// RegisterUser -
func (ps *ProviderStub) RegisterUser(accountAddress, accountTag, guardian string) ([]byte, error) {
	if ps.RegisterUserCalled != nil {
		return ps.RegisterUserCalled(accountAddress, accountTag, guardian)
	}
	return make([]byte, 0), nil
}

// IsInterfaceNil -
func (ps *ProviderStub) IsInterfaceNil() bool {
	return ps == nil
}
