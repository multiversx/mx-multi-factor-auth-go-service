package providers

import (
	"crypto"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const (
	minBackoff     = time.Second
	minMaxFailures = 1
)

// ArgTimeBasedOneTimePassword is the DTO used to create a new instance of timeBasedOnetimePassword
type ArgTimeBasedOneTimePassword struct {
	TOTPHandler       handlers.TOTPHandler
	OTPStorageHandler handlers.OTPStorageHandler
	MaxFailures       uint64
	BackoffTime       time.Duration
}

type timeBasedOnetimePassword struct {
	sync.Mutex
	totpHandler               handlers.TOTPHandler
	storageOTPHandler         handlers.OTPStorageHandler
	totalVerificationFailures map[string]uint64
	frozenUsers               map[string]time.Time
	maxFailures               uint64
	backoffTime               time.Duration
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
		maxFailures:               args.MaxFailures,
		backoffTime:               args.BackoffTime,
	}, nil
}

func checkArgs(args ArgTimeBasedOneTimePassword) error {
	if check.IfNil(args.TOTPHandler) {
		return ErrNilTOTPHandler
	}
	if check.IfNil(args.OTPStorageHandler) {
		return ErrNilStorageHandler
	}
	if args.BackoffTime < minBackoff {
		return fmt.Errorf("%w for BackoffTime, received %d, min expected %d", ErrInvalidValue, args.BackoffTime, minBackoff)
	}
	if args.MaxFailures < minMaxFailures {
		return fmt.Errorf("%w for MaxFailures, received %d, min expected %d", ErrInvalidValue, args.MaxFailures, minMaxFailures)
	}

	return nil
}

// ValidateCode will validate the code provided by the user
func (totp *timeBasedOnetimePassword) ValidateCode(account, guardian []byte, userIp string, userCode string) error {
	otp, err := totp.storageOTPHandler.Get(account, guardian)
	if err != nil {
		return err
	}

	isAllowed := totp.isVerificationAllowed(account, userIp)
	if !isAllowed {
		return ErrLockDown
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
	key := computeVerificationKey(account, ip)

	totp.Lock()
	defer totp.Unlock()

	totp.totalVerificationFailures[key]++
	if totp.totalVerificationFailures[key] >= totp.maxFailures {
		delete(totp.totalVerificationFailures, key)
		totp.frozenUsers[key] = time.Now()
	}
}

func (totp *timeBasedOnetimePassword) isVerificationAllowed(account []byte, ip string) bool {
	key := computeVerificationKey(account, ip)

	totp.Lock()
	defer totp.Unlock()

	frozenTime, found := totp.frozenUsers[key]
	if !found {
		return true
	}

	if totp.isBackoffExpired(frozenTime) {
		delete(totp.frozenUsers, key)
		return true
	}

	return false
}

// Checks the time difference between the function call time and the parameter
// if the difference of time is greater than BACKOFF_MINUTES  it returns true, otherwise false
func (totp *timeBasedOnetimePassword) isBackoffExpired(backoffStartTime time.Time) bool {
	backoffEndingTime := backoffStartTime.UTC().Add(totp.backoffTime)
	return time.Now().UTC().After(backoffEndingTime)
}

// IsInterfaceNil returns true if there is no value under the interface
func (totp *timeBasedOnetimePassword) IsInterfaceNil() bool {
	return totp == nil
}

func computeVerificationKey(account []byte, ip string) string {
	return string(account) + ":" + ip
}
