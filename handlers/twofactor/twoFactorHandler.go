package twofactor

import (
	"crypto"

	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// twoFactorHandler is a wrapper over two factor totp implementation
type twoFactorHandler struct {
	otpProvider handlers.OTPProvider
	hashType    crypto.Hash
}

// NewTwoFactorHandler returns a new instance of twoFactorHandler
func NewTwoFactorHandler(otpProvider handlers.OTPProvider, hashType crypto.Hash) (*twoFactorHandler, error) {
	if check.IfNil(otpProvider) {
		return nil, handlers.ErrNilOTPProvider
	}

	return &twoFactorHandler{
		otpProvider: otpProvider,
		hashType:    hashType,
	}, nil
}

// CreateTOTP returns a new two factor totp
func (handler *twoFactorHandler) CreateTOTP(account string) (handlers.OTP, error) {
	return handler.otpProvider.GenerateTOTP(account, handler.hashType)
}

// TOTPFromBytes returns a two factor totp from bytes
func (handler *twoFactorHandler) TOTPFromBytes(encryptedMessage []byte) (handlers.OTP, error) {
	return handler.otpProvider.TOTPFromBytes(encryptedMessage)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *twoFactorHandler) IsInterfaceNil() bool {
	return handler == nil
}
