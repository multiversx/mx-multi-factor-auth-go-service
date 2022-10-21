package providers

import (
	"crypto"
	"fmt"
	"sync"
)

const otpsEncodedFileName = "otpsEncoded"

type timebasedOnetimePassword struct {
	issuer      string
	digits      int
	otps        map[string]Totp
	otpsEncoded map[string][]byte
	sync.RWMutex

	createNewOtpHandle  func(account, issuer string, hash crypto.Hash, digits int) (Totp, error)
	totpFromBytesHandle func(encryptedMessage []byte, issuer string) (Totp, error)
	readOtpsHandle      func(filename string) (map[string][]byte, error)
	saveOtpHandle       func(filename string, otps map[string][]byte) error
}

// NewTimebasedOnetimePassword returns a new instance of timebasedOnetimePassword
func NewTimebasedOnetimePassword(issuer string, digits int) *timebasedOnetimePassword {
	return &timebasedOnetimePassword{
		issuer:              issuer,
		digits:              digits,
		otps:                make(map[string]Totp),
		otpsEncoded:         make(map[string][]byte),
		createNewOtpHandle:  newTOTP,
		totpFromBytesHandle: totpFromBytes,
		readOtpsHandle:      readOtps,
		saveOtpHandle:       saveOtp,
	}
}

// LoadSavedAccounts will load the otps saved
func (p *timebasedOnetimePassword) LoadSavedAccounts() error {
	otpsEncoded, err := p.readOtpsHandle(otpsEncodedFileName)
	if err != nil {
		return err
	}
	if otpsEncoded == nil {
		otpsEncoded = make(map[string][]byte)
	}
	otps := make(map[string]Totp)

	for address, otp := range otpsEncoded {
		otps[address], err = p.totpFromBytesHandle(otp, p.issuer)
		if err != nil {
			return err
		}
	}
	return nil
}

// VerifyCodeAndUpdateOTP will validate the code provided by the user
func (p *timebasedOnetimePassword) VerifyCodeAndUpdateOTP(account, userCode string) error {
	otp, exists := p.otps[account]
	if !exists {
		return fmt.Errorf("%w: %s", ErrNoOtpForAddress, account)
	}

	errValidation := otp.Validate(userCode)
	err := p.updateIfNeeded(account, otp)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCannotUpdateInformation, err)
	}
	if errValidation != nil {
		return fmt.Errorf("%w: %s", ErrInvalidCode, errValidation)
	}

	return nil
}

// RegisterUser generates a new timebasedOnetimePassword returning the QR code required for user to set up the OTP on his end
func (p *timebasedOnetimePassword) RegisterUser(account string) ([]byte, error) {
	// TODO: check that the user actually has the sk of the address
	otp, err := p.createNewOtpHandle(account, p.issuer, crypto.SHA1, p.digits)
	if err != nil {
		return nil, err
	}

	qrBytes, err := otp.QR()
	if err != nil {
		return nil, err
	}

	err = p.updateIfNeeded(account, otp)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCannotUpdateInformation, err)
	}

	return qrBytes, nil
}

func (p *timebasedOnetimePassword) updateIfNeeded(account string, otp Totp) error {
	p.Lock()
	defer p.Unlock()

	otpBytes, err := otp.ToBytes()
	if err != nil {
		return err
	}

	oldOtpEncoded, exists := p.otpsEncoded[account]
	isSameOtp := string(otpBytes) == string(oldOtpEncoded)
	if exists && isSameOtp {
		return nil
	}

	p.otpsEncoded[account] = otpBytes
	err = p.saveOtpHandle(otpsEncodedFileName, p.otpsEncoded)
	if err != nil {
		if exists {
			p.otpsEncoded[account] = oldOtpEncoded
		}
		return err
	}
	p.otps[account] = otp

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *timebasedOnetimePassword) IsInterfaceNil() bool {
	return p == nil
}
