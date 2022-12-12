package providers

// Provider defines the actions needed to be performed by a multi-auth provider
type Provider interface {
	ValidateCode(account, guardian, userCode string) error
	RegisterUser(accountAddress, accountTag, guardian string) ([]byte, error)
	IsInterfaceNil() bool
}
