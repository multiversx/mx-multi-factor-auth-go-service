package providers

import (
	"crypto"
	"sync"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const (
	backoff_minutes = 5 // this is the time to wait before verifying another token
	max_failures    = 3 // total amount of failures, after that the user needs to wait for the backoff time
)

// ArgTimeBasedOneTimePassword is the DTO used to create a new instance of timeBasedOnetimePassword
type ArgTimeBasedOneTimePassword struct {
	TOTPHandler       handlers.TOTPHandler
	OTPStorageHandler handlers.OTPStorageHandler
}

type timeBasedOnetimePassword struct {
	sync.Mutex
	totpHandler               handlers.TOTPHandler
	storageOTPHandler         handlers.OTPStorageHandler
	totalVerificationFailures map[string]uint64
	frozenUsers               map[string]time.Time
}

// NewTimeBasedOnetimePassword returns a new instance of timeBasedOnetimePassword
func NewTimeBasedOnetimePassword(args ArgTimeBasedOneTimePassword) (*timeBasedOnetimePassword, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &timeBasedOnetimePassword{
		totpHandler:               args.TOTPHandler,
		totalVerificationFailures: make(map[string]uint64),
		frozenUsers:               make(map[string]time.Time),
		storageOTPHandler:         args.OTPStorageHandler,
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
func (totp *timeBasedOnetimePassword) ValidateCode(account, guardian []byte, userIp string, userCode string) error {
	otp, err := totp.storageOTPHandler.Get(account, guardian)
	if err != nil {
		return err
	}

	isFrozen := totp.checkFrozen(account, userIp)
	if isFrozen {
		return ErrFrozenAccount
	}

	err = otp.Validate(userCode)
	if err != nil {
		totp.incrementFailures(account, userIp)
	}
	return err
}

// RegisterUser generates a new timeBasedOnetimePassword returning the QR code required for user to set up the OTP on his end
func (totp *timeBasedOnetimePassword) RegisterUser(accountAddress, guardian []byte, accountTag string) ([]byte, error) {
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

func (totp *timeBasedOnetimePassword) incrementFailures(account []byte, ip string) {
	totp.Lock()
	defer totp.Unlock()

	key := string(account) + ":" + ip
	failures, found := totp.totalVerificationFailures[key]
	if !found {
		failures = 0
	}

	totp.totalVerificationFailures[key] = failures + 1
	if totp.totalVerificationFailures[key] >= max_failures {
		delete(totp.totalVerificationFailures, key)
		totp.frozenUsers[key] = time.Now()
	}
}

func (totp *timeBasedOnetimePassword) checkFrozen(account []byte, ip string) bool {
	totp.Lock()
	defer totp.Unlock()

	key := string(account) + ":" + ip

	frozenTime, found := totp.frozenUsers[key]
	if !found {
		return false
	}

	if validBackoffTime(frozenTime) {
		delete(totp.frozenUsers, key)
		return false
	}

	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (totp *timeBasedOnetimePassword) IsInterfaceNil() bool {
	return totp == nil
}

// Checks the time difference between the function call time and the parameter
// if the difference of time is greater than BACKOFF_MINUTES  it returns true, otherwise false
func validBackoffTime(lastVerification time.Time) bool {
	diff := lastVerification.UTC().Add(backoff_minutes * time.Minute)
	return time.Now().UTC().After(diff)
}
