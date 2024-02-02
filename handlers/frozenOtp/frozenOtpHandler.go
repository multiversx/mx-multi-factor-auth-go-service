package frozenOtp

import (
	"math"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("FrozenOtpHandler")

// ArgsFrozenOtpHandler is the DTO used to create a new instance of frozenOtpHandler
type ArgsFrozenOtpHandler struct {
	RateLimiter redis.RateLimiter
}

type frozenOtpHandler struct {
	rateLimiter redis.RateLimiter
}

// NewFrozenOtpHandler returns a new instance of frozenOtpHandler
func NewFrozenOtpHandler(args ArgsFrozenOtpHandler) (*frozenOtpHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &frozenOtpHandler{
		rateLimiter: args.RateLimiter,
	}, nil
}

func checkArgs(args ArgsFrozenOtpHandler) error {
	if check.IfNil(args.RateLimiter) {
		return handlers.ErrNilRateLimiter
	}

	return nil
}

// BackOffTime returns the configured back off time in seconds
func (totp *frozenOtpHandler) BackOffTime() uint64 {
	return uint64(totp.rateLimiter.Period().Seconds())
}

// MaxFailures returns the configured max failures
func (totp *frozenOtpHandler) MaxFailures() uint64 {
	return uint64(totp.rateLimiter.Rate())
}

// IsVerificationAllowedAndIncreaseTrials returns true if the account and ip are not frozen, otherwise false
func (totp *frozenOtpHandler) IsVerificationAllowedAndIncreaseTrials(account string, ip string) (*requests.OTPCodeVerifyData, error) {
	key := computeVerificationKey(account, ip)

	res, err := totp.rateLimiter.CheckAllowedAndIncreaseTrials(key)
	if err != nil {
		return nil, err
	}
	verifyCodeAllowData := &requests.OTPCodeVerifyData{
		RemainingTrials: res.Remaining,
		ResetAfter:      int(math.Round(res.ResetAfter.Seconds())),
	}

	if !res.Allowed {
		log.Debug("User is now frozen",
			"address", account,
			"ip", ip,
		)

		return verifyCodeAllowData, core.ErrTooManyFailedAttempts
	}

	return verifyCodeAllowData, nil
}

// Reset removes the account and ip from local cache
func (totp *frozenOtpHandler) Reset(account string, ip string) {
	key := computeVerificationKey(account, ip)

	err := totp.rateLimiter.Reset(key)
	if err != nil {
		log.Error("failed to reset limiter for key", "key", key, "error", err.Error())
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (totp *frozenOtpHandler) IsInterfaceNil() bool {
	return totp == nil
}

func computeVerificationKey(account string, ip string) string {
	return account + ":" + ip
}
