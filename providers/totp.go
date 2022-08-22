package providers

import (
	"crypto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/sec51/twofactor"
)

const otpsEncodedFileName = "otpsEncoded"

type totp struct {
	issuer      string
	digits      int
	otps        map[string]*twofactor.Totp
	otpsEncoded map[string][]byte
	sync.RWMutex
}

// NewTOTP returns a new instance of totp
func NewTOTP(issuer string, digits int) (*totp, error) {
	otpsEncoded, err := readOtps(otpsEncodedFileName)
	if otpsEncoded == nil {
		otpsEncoded = make(map[string][]byte)
	}
	otps := make(map[string]*twofactor.Totp)

	for address, otp := range otpsEncoded {
		otps[address], err = twofactor.TOTPFromBytes(otp, issuer)
		if err != nil {
			return nil, err
		}
	}

	return &totp{
		issuer:      issuer,
		digits:      digits,
		otps:        otps,
		otpsEncoded: otpsEncoded,
	}, nil

}

// Validate will validate the code provided by the user
func (p *totp) Validate(account, userCode string) (bool, error) {
	otp, exists := p.otps[account]
	if !exists {
		return false, fmt.Errorf("no otp created for account: %s", account)
	}
	errValidation := otp.Validate(userCode)
	err := p.update(account, otp)
	if err != nil {
		return false, err
	}
	if errValidation != nil {
		return false, errValidation
	}

	isValid := err == nil
	return isValid, err
}

// RegisterUser generates a new TOTP returning the QR code required for user to set up the OTP on his end
func (p *totp) RegisterUser(account string) ([]byte, error) {
	// TODO: check that the user actually has the sk of the address
	otp, err := twofactor.NewTOTP(account, p.issuer, crypto.SHA1, p.digits)
	if err != nil {
		return nil, err
	}

	qrBytes, err := otp.QR()
	if err != nil {
		return nil, err
	}

	err = p.update(account, otp)
	if err != nil {
		return nil, err
	}

	return qrBytes, nil
}

func (p *totp) update(account string, otp *twofactor.Totp) error {
	p.Lock()
	defer p.Unlock()
	otpBytes, err := otp.ToBytes()
	if err != nil {
		return nil
	}
	oldOtpEncoded, exists := p.otpsEncoded[account]
	p.otpsEncoded[account] = otpBytes
	err = saveOtp(otpsEncodedFileName, p.otpsEncoded)
	if err != nil {
		if exists {
			p.otpsEncoded[account] = oldOtpEncoded
		}
		return err
	}
	p.otps[account] = otp

	return nil
}

func readOtps(filename string) (map[string][]byte, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("%s.json", filename))
	if err != nil {
		return nil, err
	}
	var otpsEncoded map[string][]byte
	err = json.Unmarshal(data, &otpsEncoded)

	return otpsEncoded, err
}

func saveOtp(filename string, otps map[string][]byte) error {
	filePath := fmt.Sprintf("%s.json", filename)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0777)
	defer file.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	jsonOtps, err := json.Marshal(otps)

	_, err = file.Write(jsonOtps)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *totp) IsInterfaceNil() bool {
	return p == nil
}
