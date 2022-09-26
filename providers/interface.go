package providers

// Totp defines the methods available for a time based one time password provider
type Totp interface {
	Validate(userCode string) error
	OTP() (string, error)
	QR() ([]byte, error)
	ToBytes() ([]byte, error)
}
