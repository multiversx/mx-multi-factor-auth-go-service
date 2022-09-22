package providers

import "crypto"

// NewTimebasedOnetimePassword returns a new instance of timebasedOnetimePassword
func NewTimebasedOnetimePasswordWithHandler(
	issuer string, digits int, createNewOtpHandle func(account, issuer string, hash crypto.Hash, digits int) (Totp, error), totpFromBytesHandle func(encryptedMessage []byte, issuer string) (Totp, error), readOtpsHandle func(filename string) (map[string][]byte, error), saveOtpHandle func(filename string, otps map[string][]byte) error) *timebasedOnetimePassword {
	return &timebasedOnetimePassword{
		issuer:              issuer,
		digits:              digits,
		otps:                make(map[string]Totp),
		otpsEncoded:         make(map[string][]byte),
		createNewOtpHandle:  createNewOtpHandle,
		totpFromBytesHandle: totpFromBytesHandle,
		readOtpsHandle:      readOtpsHandle,
		saveOtpHandle:       saveOtpHandle,
	}
}
