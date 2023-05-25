package frozenOtp

import (
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
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

// IsVerificationAllowed returns true if the account and ip are not frozen, otherwise false
func (totp *frozenOtpHandler) IsVerificationAllowed(account string, ip string) (*requests.OTPCodeVerifyData, bool) {
	key := computeVerificationKey(account, ip)

	res, err := totp.rateLimiter.CheckAllowed(key)
	if err != nil {
		return nil, false
	}
	verifyCodeAllowData := &requests.OTPCodeVerifyData{
		VerifyData: &requests.OTPCodeVerifyDataPayload{
			RemainingTrials: res.Remaining,
			ResetAfter:      int(res.ResetAfter.Seconds()),
		},
	}

	if res.Remaining == 0 {
		log.Debug("User is now frozen",
			"address", account,
			"ip", ip,
		)

		return verifyCodeAllowData, false
	}

	return verifyCodeAllowData, true
}

// Reset removes the account and ip from local cache
func (totp *frozenOtpHandler) Reset(account string, ip string) {
	key := computeVerificationKey(account, ip)

	err := totp.rateLimiter.Reset(key)
	if err != nil {
		log.Warn("failed to reset limiter for key", "key", key, "error", err.Error())
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (totp *frozenOtpHandler) IsInterfaceNil() bool {
	return totp == nil
}

func computeVerificationKey(account string, ip string) string {
	return account + ":" + ip
}
