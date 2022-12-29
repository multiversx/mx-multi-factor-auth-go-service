package providers

import (
	"crypto"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
)

// ArgTimeBasedOneTimePassword is the DTO used to create a new instance of timeBasedOnetimePassword
type ArgTimeBasedOneTimePassword struct {
	TOTPHandler       handlers.TOTPHandler
	OTPStorageHandler handlers.OTPStorageHandler
}

type timeBasedOnetimePassword struct {
	totpHandler       handlers.TOTPHandler
	storageOTPHandler handlers.OTPStorageHandler
}

// NewTimeBasedOnetimePassword returns a new instance of timeBasedOnetimePassword
func NewTimeBasedOnetimePassword(args ArgTimeBasedOneTimePassword) (*timeBasedOnetimePassword, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &timeBasedOnetimePassword{
		totpHandler:       args.TOTPHandler,
		storageOTPHandler: args.OTPStorageHandler,
	}, nil
}

func checkArgs(args ArgTimeBasedOneTimePassword) error {
	if check.IfNil(args.TOTPHandler) {
		return ErrNilTOTPHandler
	}
	if check.IfNil(args.OTPStorageHandler) {
		return ErrNilStorageHandler
	}

	return nil
}

// ValidateCode will validate the code provided by the user
func (totp *timeBasedOnetimePassword) ValidateCode(account, guardian, userCode string) error {
	otp, err := totp.storageOTPHandler.Get(account, guardian)
	if err != nil {
		return err
	}

	return otp.Validate(userCode)
}

// RegisterUser generates a new timeBasedOnetimePassword returning the QR code required for user to set up the OTP on his end
func (totp *timeBasedOnetimePassword) RegisterUser(accountAddress, accountTag, guardian string) ([]byte, error) {
	otp, err := totp.totpHandler.CreateTOTP(accountTag, crypto.SHA1)
	if err != nil {
		return nil, err
	}

	qrBytes, err := otp.QR()
	if err != nil {
		return nil, err
	}

	err = totp.storageOTPHandler.Save(accountAddress, guardian, otp)
	if err != nil {
		return nil, err
	}

	return qrBytes, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (totp *timeBasedOnetimePassword) IsInterfaceNil() bool {
	return totp == nil
}
