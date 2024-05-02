package factory

import (
	"crypto"

	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/secureOtp"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/twofactor"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/twofactor/sec51"
	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
)

const hashType = crypto.SHA1

// CreateOTPHandler will create a new otp handler instance
func CreateOTPHandler(configs *config.Configs) (handlers.TOTPHandler, error) {
	otpProvider := sec51.NewSec51Wrapper(configs.GeneralConfig.TwoFactor.Digits, configs.GeneralConfig.TwoFactor.Issuer)
	return twofactor.NewTwoFactorHandler(otpProvider, hashType)
}

// CreateSecureOTPHandler will create a new otp handler instance
func CreateSecureOTPHandler(configs *config.Configs) (handlers.SecureOtpHandler, error) {
	rateLimiter, err := redis.CreateRedisRateLimiter(configs.ExternalConfig.Redis, configs.GeneralConfig.TwoFactor)
	if err != nil {
		return nil, err
	}

	secureOtpArgs := secureOtp.ArgsSecureOtpHandler{
		RateLimiter: rateLimiter,
	}
	return secureOtp.NewSecureOtpHandler(secureOtpArgs)
}
