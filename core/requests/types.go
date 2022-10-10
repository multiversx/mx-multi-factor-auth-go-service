package requests

import "github.com/ElrondNetwork/elrond-sdk-erdgo/data"

// Code defines a pair of (provider; code)
type Code struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
}

// SendTransaction is the JSON request the service is receiving
// when a user sends a new transaction to be sign by the guardian and send
type SendTransaction struct {
	Credentials string           `json:"credentials"`
	Account     string           `json:"account"`
	Codes       []Code           `json:"codes"`
	Tx          data.Transaction `json:"transaction"`
}

// VerifyCodes represents the JSON requests a user uses to validate authentication codes
type VerifyCodes struct {
	Credentials string `json:"credentials"`
	Codes       []Code `json:"codes"`
	Guardian    string `json:"guardian"`
}

// Register represents the JSON requests a user uses to require a new provider registration
type Register struct {
	Credentials string `json:"credentials"`
	Provider    string `json:"provider"`
	Guardian    string `json:"guardian"`
}

// GetGuardianAddress represents the JSON requests a user uses to require a guardian address
type GetGuardianAddress struct {
	Credentials string `json:"credentials"`
	Provider    string `json:"provider"`
}
