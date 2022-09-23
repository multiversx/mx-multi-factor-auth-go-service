package core

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

type Guardian interface {
	ValidateAndSend(transaction data.Transaction) (string, error)
	IsInterfaceNil() bool
}

// Provider defines the actions needed to be performed by a multi-auth provider
type Provider interface {
	LoadSavedAccounts() error
	Validate(account, userCode string) (bool, error)
	RegisterUser(account string) ([]byte, error)
	IsInterfaceNil() bool
}
