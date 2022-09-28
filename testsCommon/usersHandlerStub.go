package testsCommon

// UsersHandlerStub -
type UsersHandlerStub struct {
	AddUserCalled    func(address string) error
	HasUserCalled    func(address string) bool
	RemoveUserCalled func(address string)
}

// AddUser -
func (stub *UsersHandlerStub) AddUser(address string) error {
	if stub.AddUserCalled != nil {
		return stub.AddUserCalled(address)
	}
	return nil
}

// HasUser -
func (stub *UsersHandlerStub) HasUser(address string) bool {
	if stub.HasUserCalled != nil {
		return stub.HasUserCalled(address)
	}
	return false
}

// RemoveUser -
func (stub *UsersHandlerStub) RemoveUser(address string) {
	if stub.RemoveUserCalled != nil {
		stub.RemoveUserCalled(address)
	}
}

// IsInterfaceNil -
func (stub *UsersHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
