package testsCommon

// UsersHandlerStub -
type UsersHandlerStub struct {
	AddUserCalled    func(address string)
	HasUserCalled    func(address string) bool
	RemoveUserCalled func(address string)
}

// AddUser -
func (stub *UsersHandlerStub) AddUser(address string) {
	if stub.AddUserCalled != nil {
		stub.AddUserCalled(address)
	}
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
