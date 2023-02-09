package requests

import "github.com/multiversx/mx-sdk-go/data"

// SignTransaction is the JSON request the service is receiving
// when a user sends a new transaction to be signed by the guardian
type SignTransaction struct {
	Code string           `json:"code"`
	Tx   data.Transaction `json:"transaction"`
}

// SignMultipleTransactions is the JSON request the service is receiving
// when a user sends multiple transactions to be signed by the guardian
type SignMultipleTransactions struct {
	Code string             `json:"code"`
	Txs  []data.Transaction `json:"transactions"`
}

// VerificationPayload represents the JSON requests a user uses to validate the authentication code
type VerificationPayload struct {
	Code     string `json:"code"`
	Guardian string `json:"guardian"`
}

// RegistrationPayload represents the JSON requests a user uses to require a new provider registration
type RegistrationPayload struct {
	Tag string `json:"tag"`
}

// RegisterReturnData represents the returned data for a registration request
type RegisterReturnData struct {
	QR              []byte `json:"qr"`
	GuardianAddress string `json:"guardian-address"`
}
