package integrationtests

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/frozenOtp"
	redisLocal "github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestOTPRateLimiting_FailuresBlocking(t *testing.T) {
	maxFailures := 3
	periodLimit := 9

	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	redisLimiter, err := redisLocal.NewRedisClientWrapper(redisClient, "rate:")
	require.Nil(t, err)

	rateLimiterArgs := redisLocal.ArgsRateLimiter{
		OperationTimeoutInSec: 1000,
		MaxFailures:           uint64(maxFailures),
		LimitPeriodInSec:      uint64(periodLimit),
		Storer:                redisLimiter,
	}
	rl, err := redisLocal.NewRateLimiter(rateLimiterArgs)
	require.Nil(t, err)

	frozenOtpArgs := frozenOtp.ArgsFrozenOtpHandler{
		RateLimiter: rl,
	}
	frozenOtpHandler, err := frozenOtp.NewFrozenOtpHandler(frozenOtpArgs)
	require.Nil(t, err)

	testOTPRateLimitingFailuresRate(t, frozenOtpHandler)
	testOTPRateLimitingTimeControl(t, frozenOtpHandler)
}

func testOTPRateLimitingFailuresRate(t *testing.T, frozenOtpHandler handlers.FrozenOtpHandler) {
	t.Run("should work 3 times after reset", func(t *testing.T) {
		userAddress := "addr0"
		userIp := "ip0"

		otpVerifyData, isAllowed := frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      3,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		frozenOtpHandler.Reset(userAddress, userIp)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      3,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 1,
			ResetAfter:      6,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.False(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})

	t.Run("should not work anymore after 3 trials", func(t *testing.T) {
		userAddress := "addr1"
		userIp := "ip1"

		otpVerifyData, isAllowed := frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      3,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 1,
			ResetAfter:      6,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.False(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		time.Sleep(time.Second * time.Duration(3))

		// try multiple times to make sure ResetAfter is not over increasing

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.False(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})
}

func testOTPRateLimitingTimeControl(t *testing.T, frozenOtpHandler handlers.FrozenOtpHandler) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("should work 3 times after partial time passed", func(t *testing.T) {
		userAddress := "addr2"
		userIp := "ip2"

		otpVerifyData, isAllowed := frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      3,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		time.Sleep(time.Second * time.Duration(expOtpVerifyData.ResetAfter))

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      3,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 1,
			ResetAfter:      6,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.False(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})

	t.Run("should work after full time passed", func(t *testing.T) {
		userAddress := "addr3"
		userIp := "ip3"

		otpVerifyData, isAllowed := frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      3,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 1,
			ResetAfter:      6,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.False(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		time.Sleep(time.Second * time.Duration(expOtpVerifyData.ResetAfter))

		otpVerifyData, isAllowed = frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
		require.True(t, isAllowed)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      3,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})
}
