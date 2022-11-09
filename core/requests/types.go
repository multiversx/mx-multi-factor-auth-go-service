package requests

import "github.com/ElrondNetwork/elrond-sdk-erdgo/data"

// SendTransaction is the JSON request the service is receiving
// when a user sends a new transaction to be signed by the guardian
type SendTransaction struct {
	Credentials string           `json:"credentials"`
	Code        string           `json:"code"`
	Tx          data.Transaction `json:"transaction"`
}

// SendMultipleTransaction is the JSON request the service is receiving
// when a user sends multiple transactions to be signed by the guardian
type SendMultipleTransaction struct {
	Credentials string             `json:"credentials"`
	Code        string             `json:"code"`
	Txs         []data.Transaction `json:"transactions"`
}

// VerificationPayload represents the JSON requests a user uses to validate the authentication code
type VerificationPayload struct {
	Credentials string `json:"credentials"`
	Code        string `json:"code"`
	Guardian    string `json:"guardian"`
}

// RegistrationPayload represents the JSON requests a user uses to require a new provider registration
type RegistrationPayload struct {
	Credentials string `json:"credentials"`
	Guardian    string `json:"guardian"`
}

// GetGuardianAddress represents the JSON requests a user uses to require a guardian address
// TODO replace Credentials with a proper struct when native-auth is ready
type GetGuardianAddress struct {
	Credentials string `json:"credentials"`
}
