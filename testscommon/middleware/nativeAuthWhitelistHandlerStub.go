package middleware

// NativeAuthWhitelistHandlerStub -
type NativeAuthWhitelistHandlerStub struct {
	IsWhitelistedCalled func(route string) bool
}

// IsWhitelisted -
func (stub *NativeAuthWhitelistHandlerStub) IsWhitelisted(route string) bool {
	if stub.IsWhitelistedCalled != nil {
		return stub.IsWhitelistedCalled(route)
	}
	return false
}

func (stub *NativeAuthWhitelistHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
