package handlers

import (
	"crypto"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
)

// TOTPHandler defines the methods available for a time based one time password handler
type TOTPHandler interface {
	CreateTOTP(account string) (OTP, error)
	TOTPFromBytes(encryptedMessage []byte) (OTP, error)
	IsInterfaceNil() bool
}

// SecureOtpHandler defines the methods available for a secure otp handler
type SecureOtpHandler interface {
	FreezeBackOffTime() uint64
	FreezeMaxFailures() uint64
	SecurityModeBackOffTime() uint64
	SecurityModeMaxFailures() uint64
	SetSecurityModeNoExpire(key string) error
	UnsetSecurityModeNoExpire(key string) error
	IsVerificationAllowedAndIncreaseTrials(account string, ip string) (*requests.OTPCodeVerifyData, error)
	Reset(account string, ip string)
	DecrementSecurityModeFailedTrials(account string) error
	ExtendSecurityMode(account string) error
	IsInterfaceNil() bool
}

// OTP defines the methods available for a one time password provider
type OTP interface {
	Validate(userCode string) error
	OTP() (string, error)
	QR() ([]byte, error)
	ToBytes() ([]byte, error)
	Url() (string, error)
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
