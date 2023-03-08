package providers

// Provider defines the actions needed to be performed by a multi-auth provider
type Provider interface {
	ValidateCode(account, guardian []byte, userCode string) error
	RegisterUser(accountAddress, guardian []byte, accountTag string) ([]byte, error)
	IsInterfaceNil() bool
}
