package requests

import "github.com/ElrondNetwork/elrond-sdk-erdgo/data"

// SignTransaction is the JSON request the service is receiving
// when a user sends a new transaction to be signed by the guardian
type SignTransaction struct {
	Credentials string           `json:"credentials"`
	Code        string           `json:"code"`
	Tx          data.Transaction `json:"transaction"`
}

// SignMultipleTransactions is the JSON request the service is receiving
// when a user sends multiple transactions to be signed by the guardian
type SignMultipleTransactions struct {
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
// TODO replace Credentials with a proper struct when native-auth is ready
type RegistrationPayload struct {
	Credentials string `json:"credentials"`
	Tag         string `json:"tag"`
}

// RegisterReturnData represents the returned data for a registration request
type RegisterReturnData struct {
	QR              []byte `json:"qr"`
	GuardianAddress string `json:"guardian-address"`
}
