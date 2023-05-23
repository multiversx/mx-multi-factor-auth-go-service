package frozenOtp

import (
	"fmt"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/multiversx/mx-chain-core-go/core/check"
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
	RateLimiter redis.RateLimiter
}

type frozenOtpHandler struct {
	maxFailures uint8
	backoffTime time.Duration
	rateLimiter redis.RateLimiter
}

// NewFrozenOtpHandler returns a new instance of frozenOtpHandler
func NewFrozenOtpHandler(args ArgsFrozenOtpHandler) (*frozenOtpHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &frozenOtpHandler{
		maxFailures: args.MaxFailures,
		backoffTime: args.BackoffTime,
		rateLimiter: args.RateLimiter,
	}, nil
}

func checkArgs(args ArgsFrozenOtpHandler) error {
	if args.BackoffTime < minBackoff {
		return fmt.Errorf("%w for BackoffTime, received %d, min expected %d", handlers.ErrInvalidConfig, args.BackoffTime, minBackoff)
	}
	if args.MaxFailures < minMaxFailures {
		return fmt.Errorf("%w for MaxFailures, received %d, min expected %d", handlers.ErrInvalidConfig, args.MaxFailures, minMaxFailures)
	}
	if check.IfNil(args.RateLimiter) {
		return handlers.ErrNilRateLimiter
	}

	return nil
}

// BackOffTime returns the configured back off time in seconds
func (totp *frozenOtpHandler) BackOffTime() uint64 {
	return uint64(totp.backoffTime.Seconds())
}

// IsVerificationAllowed returns true if the account and ip are not frozen, otherwise false
func (totp *frozenOtpHandler) IsVerificationAllowed(account string, ip string) bool {
	key := computeVerificationKey(account, ip)

	numRemaining, err := totp.rateLimiter.CheckAllowed(key, int(totp.maxFailures), totp.backoffTime)
	if err != nil {
		return false
	}

	if numRemaining == 0 {
		log.Debug("Freezing user",
			"address", account,
			"ip", ip,
		)

		return false
	}

	return true
}

// Reset removes the account and ip from local cache
func (totp *frozenOtpHandler) Reset(account string, ip string) {
	key := computeVerificationKey(account, ip)

	_ = totp.rateLimiter.Reset(key)
}

// IsInterfaceNil returns true if there is no value under the interface
func (totp *frozenOtpHandler) IsInterfaceNil() bool {
	return totp == nil
}

func computeVerificationKey(account string, ip string) string {
	return account + ":" + ip
}
