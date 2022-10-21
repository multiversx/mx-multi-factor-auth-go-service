package handlers

import (
	"crypto"

	"github.com/sec51/twofactor"
)

// twoFactorHandler is a wrapper over two factor totp implementation
type twoFactorHandler struct {
	digits int
	issuer string
}

// NewTwoFactorHandler returns a new instance of twoFactorHandler
func NewTwoFactorHandler(digits int, issuer string) *twoFactorHandler {
	return &twoFactorHandler{
		digits: digits,
		issuer: issuer,
	}
}

// CreateTOTP returns a new two factor totp
func (handler *twoFactorHandler) CreateTOTP(account string, hash crypto.Hash) (*twofactor.Totp, error) {
	return twofactor.NewTOTP(account, handler.issuer, hash, handler.digits)
}

// TOTPFromBytes returns a two factor totp from bytes
func (handler *twoFactorHandler) TOTPFromBytes(encryptedMessage []byte) (*twofactor.Totp, error) {
	return twofactor.TOTPFromBytes(encryptedMessage, handler.issuer)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *twoFactorHandler) IsInterfaceNil() bool {
	return handler == nil
}
