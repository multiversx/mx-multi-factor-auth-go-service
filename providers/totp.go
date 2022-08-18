package providers

import (
	"crypto"
	"fmt"
	"sync"

	"github.com/sec51/twofactor"
)

type totp struct {
	issuer string
	digits int
	otps   map[string]*twofactor.Totp //TODO: fork and use twofactor in order to change consts like backoff_minutes
	*sync.RWMutex
}

func NewTOTP(issuer string, digits int) (*totp, error) {
	return &totp{
		issuer: issuer,
		digits: digits,
		otps:   make(map[string]*twofactor.Totp),
	}, nil

}

func (p *totp) Validate(account, usercode string) (bool, error) {
	otp, exists := p.otps[account]
	if !exists {
		return false, fmt.Errorf("no otp created for account: %s", account)
	}
	isValid := otp.Validate(usercode)
	return isValid == nil, isValid
}

func (p *totp) RegisterUser(account string) ([]byte, error) {
	otp, err := twofactor.NewTOTP(account, p.issuer, crypto.SHA1, p.digits)
	if err != nil {
		return nil, err
	}

	qrBytes, err := otp.QR()
	if err != nil {
		return nil, err
	}

	p.Lock()
	p.otps[account] = otp
	p.Unlock()

	return qrBytes, nil
}
