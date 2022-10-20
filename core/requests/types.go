package requests

import "github.com/ElrondNetwork/elrond-sdk-erdgo/data"

// Code defines a pair of (provider; code)
type Code struct {
	Provider   string `json:"provider"`
	SecretCode string `json:"secretCode"`
}

// SendTransaction is the JSON request the service is receiving
// when a user sends a new transaction to be sign by the guardian and send
type SendTransaction struct {
	Account string           `json:"account"`
	Codes   []Code           `json:"codes"`
	Tx      data.Transaction `json:"transaction"`
}

// VerificationPayload represents the JSON requests a user uses to validate the authentication code
type VerificationPayload struct {
	Credentials string `json:"credentials"`
	Code        Code   `json:"code"`
	Guardian    string `json:"guardian"`
}

// RegistrationPayload represents the JSON requests a user uses to require a new provider registration
type RegistrationPayload struct {
	Credentials string `json:"credentials"`
	Provider    string `json:"provider"`
	Guardian    string `json:"guardian"`
}

// GetGuardianAddress represents the JSON requests a user uses to require a guardian address
// TODO replace Credentials with a proper struct when native-auth is ready
type GetGuardianAddress struct {
	Credentials string `json:"credentials"`
	Provider    string `json:"provider"`
}
