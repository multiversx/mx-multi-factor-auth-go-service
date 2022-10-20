package testsCommon

// ProviderStub -
type ProviderStub struct {
	LoadSavedAccountsCalled      func() error
	VerifyCodeAndUpdateOTPCalled func(account, userCode string) error
	RegisterUserCalled           func(account string) ([]byte, error)
}

// LoadSavedAccounts -
func (ps *ProviderStub) LoadSavedAccounts() error {
	if ps.LoadSavedAccountsCalled != nil {
		return ps.LoadSavedAccountsCalled()
	}
	return nil
}

// VerifyCodeAndUpdateOTP -
func (ps *ProviderStub) VerifyCodeAndUpdateOTP(account, userCode string) error {
	if ps.VerifyCodeAndUpdateOTPCalled != nil {
		return ps.VerifyCodeAndUpdateOTPCalled(account, userCode)
	}
	return nil
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
