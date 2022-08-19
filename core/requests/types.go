package requests

import "github.com/ElrondNetwork/elrond-sdk-erdgo/data"

// SendTransaction is the JSON request the service is receiving
// when an user sends a new transaction to be sign by the guardian and send
type SendTransaction struct {
	Account string `json:"account"`
	Codes   struct {
		Totp string `json:"totp"`
	} `json:"codes"`
	Tx data.Transaction `json:"transaction"`
}

// Register represents the JSON requests a user use to require a new provider registration
type Register struct {
	Account  string `json:"account"`
	Provider string `json:"provider"`
}
