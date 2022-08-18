package providers

type Provider interface {
	Validate(account, usercode string) (bool, error)
	RegisterUser(account string) ([]byte, error)
}
