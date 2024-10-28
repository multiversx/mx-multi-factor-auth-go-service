//go:generate swagger generate spec -m -o ui/swagger.json

package swagger

import (
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
)

// swagger:route GET /registered-users Guardian registeredUsers
// Returns the number of users registered.
// This request does not need the Authorization header
//
// responses:
// 200: registeredUsersResponse

// The number of users registered
// swagger:response registeredUsersResponse
type _ struct {
	// in:body
	Body struct {
		// RegisteredUsersResponse
		Data requests.RegisteredUsersResponse `json:"data"`
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:route GET /config Guardian config
// Returns the configuration values for the service instance.
// This request does not need the Authorization header
//
// responses:
// 200: configResponse

// The configuration values
// swagger:response configResponse
type _ struct {
	// in:body
	Body struct {
		// TcsConfigResponse
		Data requests.ConfigResponse `json:"data"`
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:route POST /register Guardian registerRequest
// This request is used for both new user registration and old user registration.
// A new guardian address will be returned
//
// security:
// - bearer:
// responses:
// 200: registerResponse

// Guardian address and qr code
// swagger:response registerResponse
type _ struct {
	// in:body
	Body struct {
		// RegisterReturnData
		// x-nullable:true
		Data requests.RegisterReturnData `json:"data"`
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:parameters registerRequest
type _ struct {
	// Registration payload
	// in:body
	// required:false
	Payload requests.RegistrationPayload
}

// swagger:route POST /verify-code Guardian verifyCodeRequest
// Verify code.
// Verifies the provided code against the user and guardian
//
// security:
// - bearer:
// responses:
// 400: verifyCodeResponseBadRequest
// 429: verifyCodeResponseTooManyRequests
// 200: verifyCodeResponse

// Verification result
// swagger:response verifyCodeResponse
type _ struct {
	// in:body
	Body struct {
		// Empty data field
		// x-nullable:true
		Data string `json:"data"`
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:parameters verifyCodeRequest
type _ struct {
	// Verify code payload
	// in:body
	// required:true
	Payload requests.VerificationPayload
}

// Verification result failure, bad request
// swagger:response verifyCodeResponseBadRequest
type _ struct {
	// in:body
	Body struct {
		// OTPCodeVerifyDataResponse
		Data requests.OTPCodeVerifyDataResponse `json:"data"`
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// Verification result failure, too many requests
// swagger:response verifyCodeResponseTooManyRequests
type _ struct {
	// in:body
	Body struct {
		// OTPCodeVerifyDataResponse
		Data requests.OTPCodeVerifyDataResponse `json:"data"`
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:route POST /sign-message Guardian signMessageRequest
// Sign message.
// Signs the provided message with the provided guardian
//
// responses:
// 200: signMessageResponse

// The message with the signature for it
// swagger:response signMessageResponse
type _ struct {
	// in:body
	Body struct {
		// SignMessageResponse
		// x-nullable:true
		Data requests.SignMessageResponse `json:"data"`
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:parameters signMessageRequest
type _ struct {
	// SignMessage payload
	// in:body
	// required:true
	Payload requests.SignMessage
}

// swagger:route POST /sign-transaction Guardian signTransactionRequest
// Sign transaction.
// Signs the provided transaction with the provided guardian
//
// responses:
// 200: signTransactionResponse

// The full transaction with its guardian signature on it
// swagger:response signTransactionResponse
type _ struct {
	// in:body
	Body struct {
		// SignTransactionResponse
		// x-nullable:true
		Data requests.SignTransactionResponse `json:"data"`
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:parameters signTransactionRequest
type _ struct {
	// Sign transaction payload
	// in:body
	// required:true
	Payload requests.SignTransaction
}

// swagger:route POST /sign-multiple-transactions Guardian signMultipleTransactionsRequest
// Sign multiple transactions.
// Signs the provided transactions with the provided guardian
//
// responses:
// 200: signMultipleTransactionsResponse

// The transactions array with their guardian's signature on them
// swagger:response signMultipleTransactionsResponse
type _ struct {
	// in:body
	Body struct {
		// SignMultipleTransactionsResponse
		// x-nullable:true
		Data requests.SignMultipleTransactionsResponse `json:"data"`
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:parameters signMultipleTransactionsRequest
type _ struct {
	// Sign multiple transactions payload
	// in:body
	// required:true
	Payload requests.SignMultipleTransactions
}

// swagger:route POST /set-security-mode Guardian securityModeNoExpireRequest
// Sets SecurityMode.
// Sets the security mode permanently.
// This request does not need the Authorization header
//
// responses:
// 200: setSecurityModeNoExpireMessageResponse

// The status of the operation
// swagger:response setSecurityModeNoExpireMessageResponse
type _ struct {
	// in:body
	Body struct {
		// SetSecurityModeNoExpireMessageResponse
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:parameters securityModeNoExpireRequest
type _ struct {
	// SecurityModeNoExpire payload
	// in:body
	// required:true
	Payload requests.SecurityModeNoExpire
}

// swagger:route POST /unset-security-mode Guardian unsetSecurityModeNoExpireRequest
// Unset SecurityMode.
// Unsets the security mode permanently.
// This request does not need the Authorization header
//
// responses:
// 200: unsetSecurityModeNoExpireMessageResponse

// The status of the operation
// swagger:response unsetSecurityModeNoExpireMessageResponse
type _ struct {
	// in:body
	Body struct {
		// UnsetSecurityModeNoExpireMessageResponse
		// HTTP status code
		Code string `json:"code"`
		// Internal error
		Error string `json:"error"`
	}
}

// swagger:parameters unsetSecurityModeNoExpireRequest
type _ struct {
	// UnsetSecurityModeNoExpire payload
	// in:body
	// required:true
	Payload requests.SecurityModeNoExpire
}
