package handlers

import (
	"crypto"

	"github.com/sec51/twofactor"
)

type OTPStorageHandler interface {
	Save(account, guardian string, otp OTP) error
	Get(account, guardian string) (OTP, error)
	IsInterfaceNil() bool
}

type TOTPHandler interface {
	CreateTOTP(account string, hash crypto.Hash) (*twofactor.Totp, error)
	TOTPFromBytes(encryptedMessage []byte) (*twofactor.Totp, error)
	IsInterfaceNil() bool
}

// OTP defines the methods available for a one time password provider
type OTP interface {
	Validate(userCode string) error
	OTP() (string, error)
	QR() ([]byte, error)
	ToBytes() ([]byte, error)
}
