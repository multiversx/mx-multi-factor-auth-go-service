package providers

import "crypto"

type ArgsTimebasedOnetimePasswordWithHandler struct {
	issuer              string
	digits              int
	createNewOtpHandle  func(account, issuer string, hash crypto.Hash, digits int) (Totp, error)
	totpFromBytesHandle func(encryptedMessage []byte, issuer string) (Totp, error)
	readOtpsHandle      func(filename string) (map[string][]byte, error)
	saveOtpHandle       func(filename string, otps map[string][]byte) error
}

// NewTimebasedOnetimePassword returns a new instance of timebasedOnetimePassword
func NewTimebasedOnetimePasswordWithHandler(args ArgsTimebasedOnetimePasswordWithHandler) *timebasedOnetimePassword {
	return &timebasedOnetimePassword{
		issuer:              args.issuer,
		digits:              args.digits,
		otps:                make(map[string]Totp),
		otpsEncoded:         make(map[string][]byte),
		createNewOtpHandle:  args.createNewOtpHandle,
		totpFromBytesHandle: args.totpFromBytesHandle,
		readOtpsHandle:      args.readOtpsHandle,
		saveOtpHandle:       args.saveOtpHandle,
	}
}
