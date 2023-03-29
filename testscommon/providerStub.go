package testscommon

// ProviderStub -
type ProviderStub struct {
	ValidateCodeCalled func(account, guardian []byte, userIp, userCode string) error
	RegisterUserCalled func(accountAddress, guardian []byte, accountTag string) ([]byte, error)
}

// ValidateCode -
func (ps *ProviderStub) ValidateCode(account, guardian []byte, userIp, userCode string) error {
	if ps.ValidateCodeCalled != nil {
		return ps.ValidateCodeCalled(account, guardian, userIp, userCode)
	}
	return nil
}

// RegisterUser -
func (ps *ProviderStub) RegisterUser(accountAddress, guardian []byte, accountTag string) ([]byte, error) {
	if ps.RegisterUserCalled != nil {
		return ps.RegisterUserCalled(accountAddress, guardian, accountTag)
	}
	return make([]byte, 0), nil
}

// IsInterfaceNil -
func (ps *ProviderStub) IsInterfaceNil() bool {
	return ps == nil
}
