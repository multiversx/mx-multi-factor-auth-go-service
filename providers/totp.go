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

// Validate will validate the code provided by the user
func (p *timebasedOnetimePassword) Validate(account, userCode string) (bool, error) {
	p.RLock()
	otp, exists := p.otps[account]
	p.RUnlock()
	if !exists {
		return false, fmt.Errorf("%w: %s", ErrNoOtpForAddress, account)
	}
	errValidation := otp.Validate(userCode)
	err := p.update(account, otp)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrCannotUpdateInformation, err)
	}
	if errValidation != nil {
		return false, fmt.Errorf("%w: %s", ErrInvalidCode, errValidation)
	}

	return true, nil
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

	err = p.update(account, otp)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCannotUpdateInformation, err)
	}

	return qrBytes, nil
}

// IsUserRegistered returns true if the user is registered with this provider
func (p *timebasedOnetimePassword) IsUserRegistered(account string) bool {
	p.RLock()
	defer p.RUnlock()
	_, exists := p.otps[account]
	return exists
}

func (p *timebasedOnetimePassword) update(account string, otp Totp) error {
	p.Lock()
	defer p.Unlock()
	otpBytes, err := otp.ToBytes()
	if err != nil {
		return err
	}
	oldOtpEncoded, exists := p.otpsEncoded[account]
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
