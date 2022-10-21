package providers

import (
	"crypto"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
)

const otpsEncodedFileName = "otpsEncoded"

// ArgTimeBasedOneTimePassword is the DTO used to create a new instance of timebasedOnetimePassword
type ArgTimeBasedOneTimePassword struct {
	TOTPHandler       handlers.TOTPHandler
	OTPStorageHandler handlers.OTPStorageHandler
}

type timebasedOnetimePassword struct {
	totpHandler    handlers.TOTPHandler
	fileOTPHandler handlers.OTPStorageHandler
}

// NewTimebasedOnetimePassword returns a new instance of timebasedOnetimePassword
func NewTimebasedOnetimePassword(args ArgTimeBasedOneTimePassword) (*timebasedOnetimePassword, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &timebasedOnetimePassword{
		totpHandler:    args.TOTPHandler,
		fileOTPHandler: args.OTPStorageHandler,
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

// VerifyCodeAndUpdateOTP will validate the code provided by the user
func (totp *timebasedOnetimePassword) VerifyCodeAndUpdateOTP(account, guardian, userCode string) error {
	otp, err := totp.fileOTPHandler.Get(account, guardian)
	if err != nil {
		return err
	}

	return otp.Validate(userCode)
}

// RegisterUser generates a new timebasedOnetimePassword returning the QR code required for user to set up the OTP on his end
func (totp *timebasedOnetimePassword) RegisterUser(account, guardian string) ([]byte, error) {
	// TODO: check that the user actually has the sk of the address
	otp, err := totp.totpHandler.CreateTOTP(account, crypto.SHA1)
	if err != nil {
		return nil, err
	}

	qrBytes, err := otp.QR()
	if err != nil {
		return nil, err
	}

	err = totp.fileOTPHandler.Save(account, guardian, otp)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", handlers.ErrCannotUpdateInformation, err)
	}

	return qrBytes, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (totp *timebasedOnetimePassword) IsInterfaceNil() bool {
	return totp == nil
}
