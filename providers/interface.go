package providers

type OTPHandler interface {
	SaveOTP(account, guardianAddr string)
	GetOTP(account, guardianAddr string)
}
