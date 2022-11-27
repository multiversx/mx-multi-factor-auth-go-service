package providers

import (
	"crypto"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
)

// ArgTimeBasedOneTimePassword is the DTO used to create a new instance of timebasedOnetimePassword
type ArgTimeBasedOneTimePassword struct {
	TOTPHandler       handlers.TOTPHandler
	OTPStorageHandler handlers.OTPStorageHandler
}

type timebasedOnetimePassword struct {
	totpHandler       handlers.TOTPHandler
	storageOTPHandler handlers.OTPStorageHandler
}

// NewTimebasedOnetimePassword returns a new instance of timebasedOnetimePassword
func NewTimebasedOnetimePassword(args ArgTimeBasedOneTimePassword) (*timebasedOnetimePassword, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &timebasedOnetimePassword{
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
func (totp *timebasedOnetimePassword) ValidateCode(account, guardian, userCode string) error {
	otp, err := totp.storageOTPHandler.Get(account, guardian)
	if err != nil {
		return err
	}

	return otp.Validate(userCode)
}

// RegisterUser generates a new timebasedOnetimePassword returning the QR code required for user to set up the OTP on his end
func (totp *timebasedOnetimePassword) RegisterUser(accountAddress, accountTag, guardian string) ([]byte, error) {
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
func (totp *timebasedOnetimePassword) IsInterfaceNil() bool {
	return totp == nil
}
