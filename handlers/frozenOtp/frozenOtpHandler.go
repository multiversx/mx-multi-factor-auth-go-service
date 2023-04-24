package frozenOtp

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("FrozenOtpHandler")

const (
	minBackoff     = time.Second
	minMaxFailures = 1
)

// ArgsFrozenOtpHandler is the DTO used to create a new instance of frozenOtpHandler
type ArgsFrozenOtpHandler struct {
	MaxFailures uint64
	BackoffTime time.Duration
}

type frozenOtpHandler struct {
	sync.Mutex
	totalVerificationFailures map[string]uint64
	frozenUsers               map[string]time.Time
	maxFailures               uint64
	backoffTime               time.Duration
}

// NewFrozenOtpHandler returns a new instance of frozenOtpHandler
func NewFrozenOtpHandler(args ArgsFrozenOtpHandler) (*frozenOtpHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &frozenOtpHandler{
		totalVerificationFailures: make(map[string]uint64),
		frozenUsers:               make(map[string]time.Time),
		maxFailures:               args.MaxFailures,
		backoffTime:               args.BackoffTime,
	}, nil
}

func checkArgs(args ArgsFrozenOtpHandler) error {
	if args.BackoffTime < minBackoff {
		return fmt.Errorf("%w for BackoffTime, received %d, min expected %d", handlers.ErrInvalidConfig, args.BackoffTime, minBackoff)
	}
	if args.MaxFailures < minMaxFailures {
		return fmt.Errorf("%w for MaxFailures, received %d, min expected %d", handlers.ErrInvalidConfig, args.MaxFailures, minMaxFailures)
	}

	return nil
}

// IncrementFailures increments the number of verification failures for the given account and ip
func (totp *frozenOtpHandler) IncrementFailures(account []byte, ip string) {
	key := computeVerificationKey(account, ip)

	totp.Lock()
	defer totp.Unlock()

	totp.totalVerificationFailures[key]++
	log.Debug("Incremented failures for key %s to %d", key, totp.totalVerificationFailures[key])
	if totp.totalVerificationFailures[key] >= totp.maxFailures {
		log.Debug("Freezing user %s", key)
		delete(totp.totalVerificationFailures, key)
		totp.frozenUsers[key] = time.Now()
	}
}

// IsVerificationAllowed returns true if the account and ip are not frozen, otherwise false
func (totp *frozenOtpHandler) IsVerificationAllowed(account []byte, ip string) bool {
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

	log.Debug("User %s is frozen", key)
	return false
}

// Checks the time difference between the function call time and the parameter
// if the difference of time is greater than BACKOFF_MINUTES  it returns true, otherwise false
func (totp *frozenOtpHandler) isBackoffExpired(backoffStartTime time.Time) bool {
	backoffEndingTime := backoffStartTime.UTC().Add(totp.backoffTime)
	return time.Now().UTC().After(backoffEndingTime)
}

// IsInterfaceNil returns true if there is no value under the interface
func (totp *frozenOtpHandler) IsInterfaceNil() bool {
	return totp == nil
}

func computeVerificationKey(account []byte, ip string) string {
	return string(account) + ":" + ip
}
