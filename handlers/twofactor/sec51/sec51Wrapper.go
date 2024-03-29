package sec51

import (
	"crypto"
	"errors"

	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/twofactor"
)

var emptyEncryptedMessage = errors.New("empty encrypted message")

type sec51Wrapper struct {
	digits int
	issuer string
}

// NewSec51Wrapper returns a new sec51 wrapper instance
func NewSec51Wrapper(digits int, issuer string) *sec51Wrapper {
	return &sec51Wrapper{
		digits: digits,
		issuer: issuer,
	}
}

// GenerateTOTP returns a new sec51 totp
func (s *sec51Wrapper) GenerateTOTP(account string, hash crypto.Hash) (handlers.OTP, error) {
	return twofactor.NewTOTP(account, s.issuer, hash, s.digits)
}

// TOTPFromBytes returns the totp for the provided bytes
func (s *sec51Wrapper) TOTPFromBytes(encryptedMessage []byte) (handlers.OTP, error) {
	if len(encryptedMessage) == 0 {
		return nil, emptyEncryptedMessage
	}
	return twofactor.TOTPFromBytes(encryptedMessage)
}

func (s *sec51Wrapper) IsInterfaceNil() bool {
	return s == nil
}
