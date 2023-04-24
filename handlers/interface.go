package handlers

import (
	"crypto"

	"github.com/multiversx/multi-factor-auth-go-service/core"
)

// TOTPHandler defines the methods available for a time based one time password handler
type TOTPHandler interface {
	CreateTOTP(account string) (OTP, error)
	TOTPFromBytes(encryptedMessage []byte) (OTP, error)
	IsInterfaceNil() bool
}

// FrozenOtpHandler defines the methods available for a frozen otp handler
type FrozenOtpHandler interface {
	IncrementFailures(account string, ip string)
	IsVerificationAllowed(account string, ip string) bool
	IsInterfaceNil() bool
}

// OTP defines the methods available for a one time password provider
type OTP interface {
	Validate(userCode string) error
	OTP() (string, error)
	QR() ([]byte, error)
	ToBytes() ([]byte, error)
}

// ShardedStorageFactory defines the methods available for a sharded storage factory
type ShardedStorageFactory interface {
	Create() (core.StorageWithIndex, error)
	IsInterfaceNil() bool
}

// OTPProvider defines the methods available for an otp provider
type OTPProvider interface {
	GenerateTOTP(account string, hash crypto.Hash) (OTP, error)
	TOTPFromBytes(encryptedMessage []byte) (OTP, error)
	IsInterfaceNil() bool
}
