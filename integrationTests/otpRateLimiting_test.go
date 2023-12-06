package integrationtests

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/frozenOtp"
	redisLocal "github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type miniRedisHandler interface {
	FastForward(duration time.Duration)
	Start() error
	Close()
}

func createRateLimiter(t *testing.T, maxFailures, periodLimit int) (handlers.FrozenOtpHandler, miniRedisHandler) {
	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	redisLimiter, err := redisLocal.NewRedisClientWrapper(redisClient)
	require.Nil(t, err)

	rateLimiterArgs := redisLocal.ArgsRateLimiter{
		OperationTimeoutInSec: 1000,
		MaxFailures:           int64(maxFailures),
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

	return frozenOtpHandler, server
}

func TestRateLimiter_ReconnectAfterFailure(t *testing.T) {
	maxFailures := 3
	periodLimit := 9

	frozenOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit)

	userAddress := "addr0"
	userIp := "ip0"

	_, err := frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.Nil(t, err)

	_, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.Nil(t, err)

	redisServer.Close()

	_, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.NotNil(t, err)
	require.NotEqual(t, core.ErrTooManyFailedAttempts, err)

	_ = redisServer.Start()

	_, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.Nil(t, err)

	_, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	require.Equal(t, core.ErrTooManyFailedAttempts, err)
}

func TestOTPRateLimiting_FailuresBlocking(t *testing.T) {
	maxFailures := 3
	periodLimit := 9

	t.Run("should work 3 times after reset", func(t *testing.T) {
		frozenOtpHandler, _ := createRateLimiter(t, maxFailures, periodLimit)

		userAddress := "addr0"
		userIp := "ip0"

		otpVerifyData, err := frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		frozenOtpHandler.Reset(userAddress, userIp)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 1,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})

	t.Run("should not work anymore after 3 trials", func(t *testing.T) {
		frozenOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit)

		userAddress := "addr1"
		userIp := "ip1"

		otpVerifyData, err := frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 1,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(3))

		// try multiple times to make sure ResetAfter is not over increasing

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      6,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(3))

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      3,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(3))

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 1,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(3))

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      6,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		frozenOtpHandler.Reset(userAddress, userIp)
		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      9,
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

	t.Run("should work 3 times after partial time passed", func(t *testing.T) {
		frozenOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit)

		userAddress := "addr2"
		userIp := "ip2"

		otpVerifyData, err := frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(expOtpVerifyData.ResetAfter))

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 1,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})

	t.Run("should work after full time passed", func(t *testing.T) {
		frozenOtpHandler, redisServer := createRateLimiter(t, maxFailures, periodLimit)

		userAddress := "addr3"
		userIp := "ip3"

		otpVerifyData, err := frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData := &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 1,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 0,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)

		redisServer.FastForward(time.Second * time.Duration(expOtpVerifyData.ResetAfter))

		otpVerifyData, err = frozenOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
		require.Nil(t, err)
		expOtpVerifyData = &requests.OTPCodeVerifyData{
			RemainingTrials: 2,
			ResetAfter:      9,
		}
		require.Equal(t, expOtpVerifyData, otpVerifyData)
	})
}

func TestMultipleInstanceConcurrency(t *testing.T) {
	t.Parallel()

	maxFailures := 3
	periodLimit := 9

	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	redisLimiter1, err := redisLocal.NewRedisClientWrapper(redisClient)
	require.Nil(t, err)

	rateLimiterArgs1 := redisLocal.ArgsRateLimiter{
		OperationTimeoutInSec: 1000,
		MaxFailures:           int64(maxFailures),
		LimitPeriodInSec:      uint64(periodLimit),
		Storer:                redisLimiter1,
	}
	rl1, err := redisLocal.NewRateLimiter(rateLimiterArgs1)
	require.Nil(t, err)

	redisLimiter2, err := redisLocal.NewRedisClientWrapper(redisClient)
	require.Nil(t, err)

	rateLimiterArgs2 := redisLocal.ArgsRateLimiter{
		OperationTimeoutInSec: 1000,
		MaxFailures:           int64(maxFailures),
		LimitPeriodInSec:      uint64(periodLimit),
		Storer:                redisLimiter2,
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
				_, err := rl1.CheckAllowedAndIncreaseTrials(key)
				if err == redisLocal.ErrNoExpirationTimeForKey {
					atomic.AddUint32(&cnt, 1)
				}
			case 2, 3:
				_, err := rl2.CheckAllowedAndIncreaseTrials(key)
				if err == redisLocal.ErrNoExpirationTimeForKey {
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
	assert.LessOrEqual(t, atomic.LoadUint32(&cnt), uint32(3))
}
