package requests

import (
	"time"

	"github.com/multiversx/mx-sdk-go/data"
)

// SignTransaction is the JSON request the service is receiving
// when a user sends a new transaction to be signed by the guardian
type SignTransaction struct {
	Code string           `json:"code"`
	Tx   data.Transaction `json:"transaction"`
}

// SignTransactionResponse is the service response to the sign transaction request
type SignTransactionResponse struct {
	Tx data.Transaction `json:"transaction"`
}

// SignMultipleTransactions is the JSON request the service is receiving
// when a user sends multiple transactions to be signed by the guardian
type SignMultipleTransactions struct {
	Code string             `json:"code"`
	Txs  []data.Transaction `json:"transactions"`
}

// SignMultipleTransactionsResponse is the service response to the sign multiple transactions request
type SignMultipleTransactionsResponse struct {
	Txs []data.Transaction `json:"transactions"`
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

// OTPCodeVerifyDataResponse defines the reponse data for otp code verify info
type OTPCodeVerifyDataResponse struct {
	VerifyData *OTPCodeVerifyData `json:"verification-retry-info"`
}

// OTPCodeVerifyData defines the data provided for otp code info
type OTPCodeVerifyData struct {
	RemainingTrials int `json:"remaining-trials"`
	ResetAfter      int `json:"reset-after"`
}

// DefaultOTPCodeVerifyData defines the default values for otp verify data
func DefaultOTPCodeVerifyData() *OTPCodeVerifyData {
	return &OTPCodeVerifyData{
		RemainingTrials: -1,
		ResetAfter:      -1,
	}
}

// RegisteredUsersResponse is the service response to the registered users request
type RegisteredUsersResponse struct {
	Count uint32 `json:"count"`
}

// ConfigResponse is the service response to the tcs config request
type ConfigResponse struct {
	// the minimum delay allowed between registration requests for the same guardian, in seconds
	RegistrationDelay uint32 `json:"registration-delay"`
	// the total time a user gets banned for failing too many verify code requests, in seconds
	BackoffWrongCode uint32 `json:"backoff-wrong-code"`
}

// EndpointMetricsResponse defines the response for status metrics endpoint
type EndpointMetricsResponse struct {
	NumRequests       uint64         `json:"num_requests"`
	NumTotalErrors    uint64         `json:"num_total_errors"`
	ErrorsCount       map[int]uint64 `json:"errors_count"`
	TotalResponseTime time.Duration  `json:"total_response_time"`
}
