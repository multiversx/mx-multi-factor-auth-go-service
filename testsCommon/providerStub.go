package testsCommon

// ProviderStub -
type ProviderStub struct {
	LoadSavedAccountsCalled func() error
	ValidateCalled          func(account, userCode string) (bool, error)
	RegisterUserCalled      func(account string) ([]byte, error)
}

// LoadSavedAccounts -
func (ps *ProviderStub) LoadSavedAccounts() error {
	if ps.LoadSavedAccountsCalled != nil {
		return ps.LoadSavedAccountsCalled()
	}
	return nil
}

// Validate -
func (ps *ProviderStub) Validate(account, userCode string) (bool, error) {
	if ps.ValidateCalled != nil {
		return ps.ValidateCalled(account, userCode)
	}
	return false, nil
}

// RegisterUser -
func (ps *ProviderStub) RegisterUser(account string) ([]byte, error) {
	if ps.RegisterUserCalled != nil {
		return ps.RegisterUserCalled(account)
	}
	return make([]byte, 0), nil
}

// IsInterfaceNil -
func (ps *ProviderStub) IsInterfaceNil() bool {
	return ps == nil
}
