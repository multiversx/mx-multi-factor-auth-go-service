package secureOtp

import (
	"math"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
)

var log = logger.GetOrCreate("SecureOtpHandler")

// ArgsSecureOtpHandler is the DTO used to create a new instance of secureOtpHandler
type ArgsSecureOtpHandler struct {
	RateLimiter redis.RateLimiter
}

type secureOtpHandler struct {
	rateLimiter redis.RateLimiter
}

// NewSecureOtpHandler returns a new instance of secureOtpHandler
func NewSecureOtpHandler(args ArgsSecureOtpHandler) (*secureOtpHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &secureOtpHandler{
		rateLimiter: args.RateLimiter,
	}, nil
}

func checkArgs(args ArgsSecureOtpHandler) error {
	if check.IfNil(args.RateLimiter) {
		return handlers.ErrNilRateLimiter
	}

	return nil
}

// FreezeBackOffTime returns the configured back off time in seconds for freeze
func (totp *secureOtpHandler) FreezeBackOffTime() uint64 {
	return uint64(totp.rateLimiter.Period(redis.NormalMode).Seconds())
}

// FreezeMaxFailures returns the configured max failures for freeze
func (totp *secureOtpHandler) FreezeMaxFailures() uint64 {
	return uint64(totp.rateLimiter.Rate(redis.NormalMode))
}

// SecurityModeBackOffTime returns the configured back off time in seconds in security mode
func (totp *secureOtpHandler) SecurityModeBackOffTime() uint64 {
	return uint64(totp.rateLimiter.Period(redis.SecurityMode).Seconds())
}

// SecurityModeMaxFailures returns the configured max failures in security mode
func (totp *secureOtpHandler) SecurityModeMaxFailures() uint64 {
	return uint64(totp.rateLimiter.Rate(redis.SecurityMode))
}

// IsVerificationAllowedAndIncreaseTrials returns information about the account OTP historical data, if the account and ip are not frozen or if the account has security mode activated
func (totp *secureOtpHandler) IsVerificationAllowedAndIncreaseTrials(account string, ip string) (*requests.OTPCodeVerifyData, error) {
	key := computeVerificationKey(account, ip)

	res, err := totp.rateLimiter.CheckAllowedAndIncreaseTrials(key, redis.NormalMode)
	if err != nil {
		return nil, err
	}

	// the key is the account for security mode, as this is a per user account setting
	securityModeResult, err := totp.rateLimiter.CheckAllowedAndIncreaseTrials(account, redis.SecurityMode)
	if err != nil {
		return nil, err
	}

	verifyCodeAllowData := &requests.OTPCodeVerifyData{
		RemainingTrials:             res.Remaining,
		ResetAfter:                  int(math.Round(res.ResetAfter.Seconds())),
		SecurityModeRemainingTrials: securityModeResult.Remaining,
		SecurityModeResetAfter:      int(math.Round(securityModeResult.ResetAfter.Seconds())),
	}

	if !res.Allowed {
		log.Debug("User is now frozen",
			"address", account,
			"ip", ip,
		)

		err = core.ErrTooManyFailedAttempts
	}

	if !securityModeResult.Allowed {
		log.Debug("User is now in security mode",
			"address", account,
		)
		// no need to return error, as the second code needs to be verified first
	}

	return verifyCodeAllowData, err
}

// SetSecurityModeNoExpire sets the security mode with no expire time
func (totp *secureOtpHandler) SetSecurityModeNoExpire(key string) error {
	return totp.rateLimiter.SetSecurityModeNoExpire(key)
}

// UnsetSecurityModeNoExpire unsets the security mode from persistent to volatile
func (totp *secureOtpHandler) UnsetSecurityModeNoExpire(key string) error {
	return totp.rateLimiter.UnsetSecurityModeNoExpire(key)
}

// Reset removes the account and ip from local cache
func (totp *secureOtpHandler) Reset(account string, ip string) {
	key := computeVerificationKey(account, ip)

	err := totp.rateLimiter.Reset(key)
	if err != nil {
		log.Error("failed to reset limiter for key", "key", key, "error", err.Error())
	}
}

// DecrementSecurityModeFailedTrials decrements the security mode failed trials
func (totp *secureOtpHandler) DecrementSecurityModeFailedTrials(account string) error {
	return totp.rateLimiter.DecrementSecurityFailedTrials(account)
}

// ExtendSecurityMode extends the security mode to the maximum limit
func (totp *secureOtpHandler) ExtendSecurityMode(account string) error {
	return totp.rateLimiter.ExtendSecurityMode(account)
}

// IsInterfaceNil returns true if there is no value under the interface
func (totp *secureOtpHandler) IsInterfaceNil() bool {
	return totp == nil
}

func computeVerificationKey(account string, ip string) string {
	return account + ":" + ip
}
