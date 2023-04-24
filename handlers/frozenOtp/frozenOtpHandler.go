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
	MaxFailures uint8
	BackoffTime time.Duration
}

type failuresInfo struct {
	failures     uint8
	lastFailTime time.Time
}

type frozenOtpHandler struct {
	sync.Mutex
	verificationFailures map[string]*failuresInfo
	maxFailures          uint8
	backoffTime          time.Duration
}

// NewFrozenOtpHandler returns a new instance of frozenOtpHandler
func NewFrozenOtpHandler(args ArgsFrozenOtpHandler) (*frozenOtpHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &frozenOtpHandler{
		verificationFailures: make(map[string]*failuresInfo),
		maxFailures:          args.MaxFailures,
		backoffTime:          args.BackoffTime,
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
func (totp *frozenOtpHandler) IncrementFailures(account string, ip string) {
	key := computeVerificationKey(account, ip)

	totp.Lock()
	defer totp.Unlock()

	info, found := totp.verificationFailures[key]
	if !found {
		info = &failuresInfo{}
	}

	if totp.isBackoffExpired(info.lastFailTime) {
		info.failures = 0
	}

	if info.failures >= totp.maxFailures {
		log.Debug("Freezing user",
			"address", account,
			"ip", ip,
		)

		return
	}

	info.failures++
	info.lastFailTime = time.Now()
	totp.verificationFailures[key] = info

	log.Debug("Incremented failures",
		"failures", info.failures,
		"address", account,
		"ip", ip,
	)

}

// IsVerificationAllowed returns true if the account and ip are not frozen, otherwise false
func (totp *frozenOtpHandler) IsVerificationAllowed(account string, ip string) bool {
	key := computeVerificationKey(account, ip)

	totp.Lock()
	defer totp.Unlock()

	info, found := totp.verificationFailures[key]
	if !found {
		return true
	}

	if info.failures < totp.maxFailures {
		return true
	}

	if totp.isBackoffExpired(info.lastFailTime) {
		delete(totp.verificationFailures, key)
		return true
	}

	log.Debug("User is frozen", "address", account, "ip", ip)
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

func computeVerificationKey(account string, ip string) string {
	return account + ":" + ip
}
