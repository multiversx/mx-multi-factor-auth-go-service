package integrationtests

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/secureOtp"
	redisLocal "github.com/multiversx/mx-multi-factor-auth-go-service/redis"
)

type miniRedisHandler interface {
	FastForward(duration time.Duration)
	Start() error
	Close()
}

func createRateLimiter(t *testing.T, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit int) (handlers.SecureOtpHandler, miniRedisHandler) {
	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	redisLimiter, err := redisLocal.NewRedisClientWrapper(redisClient)
	require.Nil(t, err)

	rateLimiterArgs := redisLocal.ArgsRateLimiter{
		OperationTimeoutInSec: 1000,
		FreezeFailureConfig: redisLocal.FailureConfig{
			MaxFailures:      int64(maxFailures),
			LimitPeriodInSec: uint64(periodLimit),
		},
		SecurityModeFailureConfig: redisLocal.FailureConfig{
			MaxFailures:      int64(securityModeMaxFailures),
			LimitPeriodInSec: uint64(securityModePeriodLimit),
		},
		Storer: redisLimiter,
	}
	rl, err := redisLocal.NewRateLimiter(rateLimiterArgs)
	require.Nil(t, err)

	secureOtpArgs := secureOtp.ArgsSecureOtpHandler{
		RateLimiter: rl,
	}
	secureOtpHandler, err := secureOtp.NewSecureOtpHandler(secureOtpArgs)
	require.Nil(t, err)

	return secureOtpHandler, server
}

func TestRateLimiter_ReconnectAfterFailure(t *testing.T) {
	maxFailures := 3
	periodLimit := 9
	securityModeMaxFailures := 100
	securityModePeriodLimit := 86400

	secureOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

	userAddress := "addr0"
	userIp := "ip0"

	_, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.Nil(t, err)

	_, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.Nil(t, err)

	redisServer.Close()

	_, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.NotNil(t, err)
	require.NotEqual(t, core.ErrTooManyFailedAttempts, err)

	_ = redisServer.Start()

	_, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.Nil(t, err)

	_, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.Equal(t, core.ErrTooManyFailedAttempts, err)
}

func TestOTPRateLimiting_FailuresBlocking(t *testing.T) {
	maxFailures := 3
	periodLimit := 9
	securityModeMaxFailures := 100
	securityModePeriodLimit := 86400

	t.Run("should work 3 times after reset", func(t *testing.T) {
		secureOtpHandler, _ := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

		userAddress := "addr0"
		userIp := "ip0"

		otpVerifyData, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 99,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
		secureOtpHandler.Reset(userAddress, userIp)
		err = secureOtpHandler.DecrementSecurityModeFailedTrials(userAddress)
		require.Nil(t, err)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 99,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             1,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 98,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 97,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 96,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 95,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})

	t.Run("should not work anymore after 3 trials", func(t *testing.T) {
		secureOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

		userAddress := "addr1"
		userIp := "ip1"

		otpVerifyData, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 99,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             1,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 98,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 97,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 96,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(3))

		// try multiple times to make sure ResetAfter is not over increasing
		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  6,
			SecurityModeRemainingTrials: 95,
			SecurityModeResetAfter:      86397,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(3))

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 94,
			SecurityModeResetAfter:      86394,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(3))

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 93,
			SecurityModeResetAfter:      86391,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             1,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 92,
			SecurityModeResetAfter:      86391,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(3))

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  6,
			SecurityModeRemainingTrials: 91,
			SecurityModeResetAfter:      86388,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		secureOtpHandler.Reset(userAddress, userIp)
		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 90,
			SecurityModeResetAfter:      86388,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})
}

func TestOTPRateLimiting_TimeControl(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxFailures := 3
	periodLimit := 9
	securityModeMaxFailures := 100
	securityModePeriodLimit := 86400

	t.Run("should work 3 times after partial time passed", func(t *testing.T) {
		secureOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

		userAddress := "addr2"
		userIp := "ip2"

		otpVerifyData, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 99,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(expOtpVerifyData.ResetAfter))

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 98,
			SecurityModeResetAfter:      86391,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             1,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 97,
			SecurityModeResetAfter:      86391,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 96,
			SecurityModeResetAfter:      86391,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 95,
			SecurityModeResetAfter:      86391,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})

	t.Run("should work after full time passed", func(t *testing.T) {
		secureOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

		userAddress := "addr3"
		userIp := "ip3"

		otpVerifyData, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 99,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             1,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 98,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 97,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 96,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(expOtpVerifyData.ResetAfter))

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  9,
			SecurityModeRemainingTrials: 95,
			SecurityModeResetAfter:      86391,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})
}

func TestSecurityMode(t *testing.T) {
	t.Parallel()

	maxFailures := 3
	periodLimit := 3
	securityModeMaxFailures := 3
	securityModePeriodLimit := 86400

	t.Run("test set security mode", func(t *testing.T) {

		secureOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

		userAddress := "addr2"
		userIp := "ip2"

		otpVerifyData, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 2,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		err = secureOtpHandler.SetSecurityModeNoExpire(userAddress)
		require.Nil(t, err)

		redisServer.FastForward(time.Second * time.Duration(expOtpVerifyData.ResetAfter))
		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 0,
			SecurityModeResetAfter:      -1,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(expOtpVerifyData.ResetAfter))
		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 0,
			SecurityModeResetAfter:      -1,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

	})

	t.Run("test unset when security mode is activated by user", func(t *testing.T) {
		secureOtpHandler, _ := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

		userAddress := "addr2"
		userIp := "ip2"

		otpVerifyData, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 2,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		err = secureOtpHandler.SetSecurityModeNoExpire(userAddress)
		require.Nil(t, err)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             1,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 0,
			SecurityModeResetAfter:      -1,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		err = secureOtpHandler.UnsetSecurityModeNoExpire(userAddress)
		require.Nil(t, err)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 1,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

	})

	t.Run("test unset when security mode is already activated ", func(t *testing.T) {
		secureOtpHandler, _ := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

		userAddress := "addr2"
		userIp := "ip2"

		otpVerifyData, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 2,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             1,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 1,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 0,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.NotNil(t, otpVerifyData)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)

		err = secureOtpHandler.UnsetSecurityModeNoExpire(userAddress)
		require.Nil(t, err)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.NotNil(t, otpVerifyData)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)

	})

	t.Run("test set multiple times ", func(t *testing.T) {
		secureOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

		userAddress := "addr2"
		userIp := "ip2"

		otpVerifyData, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 2,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		err = secureOtpHandler.SetSecurityModeNoExpire(userAddress)
		require.Nil(t, err)

		redisServer.FastForward(time.Second * time.Duration(expOtpVerifyData.ResetAfter))
		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 0,
			SecurityModeResetAfter:      -1,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		err = secureOtpHandler.SetSecurityModeNoExpire(userAddress)
		require.Nil(t, err)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             1,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 0,
			SecurityModeResetAfter:      -1,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})

	t.Run("test unset multiple times", func(t *testing.T) {
		secureOtpHandler, _ := createRateLimiter(t, maxFailures, periodLimit, securityModeMaxFailures, securityModePeriodLimit)

		userAddress := "addr2"
		userIp := "ip2"

		otpVerifyData, err := secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 2,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		err = secureOtpHandler.SetSecurityModeNoExpire(userAddress)
		require.Nil(t, err)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             1,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 0,
			SecurityModeResetAfter:      -1,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		err = secureOtpHandler.UnsetSecurityModeNoExpire(userAddress)
		require.Nil(t, err)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 1,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		err = secureOtpHandler.UnsetSecurityModeNoExpire(userAddress)
		require.Nil(t, err)

		otpVerifyData, err = secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials:             0,
			ResetAfter:                  3,
			SecurityModeRemainingTrials: 0,
			SecurityModeResetAfter:      86400,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

	})
}

func TestMultipleInstanceConcurrency(t *testing.T) {
	t.Parallel()

	maxFailures := 3
	periodLimit := 9
	securityModeMaxFailures := 100
	securityModePeriodLimit := 86400

	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	redisLimiter1, err := redisLocal.NewRedisClientWrapper(redisClient)
	require.Nil(t, err)

	rateLimiterArgs1 := redisLocal.ArgsRateLimiter{
		OperationTimeoutInSec: 1000,
		FreezeFailureConfig: redisLocal.FailureConfig{
			MaxFailures:      int64(maxFailures),
			LimitPeriodInSec: uint64(periodLimit),
		},
		SecurityModeFailureConfig: redisLocal.FailureConfig{
			MaxFailures:      int64(securityModeMaxFailures),
			LimitPeriodInSec: uint64(securityModePeriodLimit),
		},
		Storer: redisLimiter1,
	}
	rl1, err := redisLocal.NewRateLimiter(rateLimiterArgs1)
	require.Nil(t, err)

	redisLimiter2, err := redisLocal.NewRedisClientWrapper(redisClient)
	require.Nil(t, err)

	rateLimiterArgs2 := redisLocal.ArgsRateLimiter{
		OperationTimeoutInSec: 1000,
		FreezeFailureConfig: redisLocal.FailureConfig{
			MaxFailures:      int64(maxFailures),
			LimitPeriodInSec: uint64(periodLimit),
		},

		SecurityModeFailureConfig: redisLocal.FailureConfig{
			MaxFailures:      int64(securityModeMaxFailures),
			LimitPeriodInSec: uint64(securityModePeriodLimit),
		},
		Storer: redisLimiter2,
	}
	rl2, err := redisLocal.NewRateLimiter(rateLimiterArgs2)
	require.Nil(t, err)

	numOps := 50000
	wg := sync.WaitGroup{}

	wg.Add(numOps)

	cnt := uint32(0)
	key := "key1"
	for i := 0; i < numOps; i++ {
		go func(idx int) {
			switch idx % 6 {
			case 0, 1:
				_, err := rl1.CheckAllowedAndIncreaseTrials(key, redisLocal.NormalMode)
				if errors.Is(err, redisLocal.ErrNoExpirationTimeForKey) {
					atomic.AddUint32(&cnt, 1)
				}
			case 2, 3:
				_, err := rl2.CheckAllowedAndIncreaseTrials(key, redisLocal.NormalMode)
				if errors.Is(err, redisLocal.ErrNoExpirationTimeForKey) {
					atomic.AddUint32(&cnt, 1)
				}
			case 4:
				_ = rl1.Reset(key)
			case 5:
				_ = rl2.Reset(key)
			default:
				assert.Fail(t, "should have not been called")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	// Allow max 3 failures. This edge case may happen, but next call should be ok
	assert.LessOrEqual(t, atomic.LoadUint32(&cnt), uint32(3))
}
