package secureOtp_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/secureOtp"
	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon"
)

const (
	account = "test_account"
	ip      = "127.0.0.1"
)

var expectedErr = errors.New("expected error")

func createMockArgsSecureOtpHandler() secureOtp.ArgsSecureOtpHandler {
	return secureOtp.ArgsSecureOtpHandler{
		RateLimiter: &testscommon.RateLimiterStub{},
	}
}

func TestNewSecureOtpHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil rate limiter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSecureOtpHandler()
		args.RateLimiter = nil

		totp, err := secureOtp.NewSecureOtpHandler(args)
		require.True(t, errors.Is(err, handlers.ErrNilRateLimiter))
		require.Nil(t, totp)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSecureOtpHandler()
		totp, err := secureOtp.NewSecureOtpHandler(args)
		require.Nil(t, err)
		require.NotNil(t, totp)
		require.False(t, totp.IsInterfaceNil())
	})
}

func TestSecureOtpHandler_IsVerificationAllowedAndIncreaseTrials(t *testing.T) {
	t.Parallel()

	t.Run("on error should return err", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSecureOtpHandler()

		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedAndIncreaseTrialsCalled: func(key string, _ redis.Mode) (*redis.RateLimiterResult, error) {
				return &redis.RateLimiterResult{}, expectedErr
			},
		}
		totp, _ := secureOtp.NewSecureOtpHandler(args)

		_, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Equal(t, expectedErr, err)
	})
	t.Run("on security mode limiter check error, should return error", func(t *testing.T) {
		args := createMockArgsSecureOtpHandler()

		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedAndIncreaseTrialsCalled: func(key string, mode redis.Mode) (*redis.RateLimiterResult, error) {
				switch mode {
				case redis.NormalMode:
					return &redis.RateLimiterResult{
						Allowed:    true,
						Remaining:  3,
						ResetAfter: time.Second,
					}, nil
				case redis.SecurityMode:
					return &redis.RateLimiterResult{}, expectedErr
				default:
					return nil, errors.New("unexpected mode")
				}
			},
		}
		totp, _ := secureOtp.NewSecureOtpHandler(args)

		verifyCodeData, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Equal(t, expectedErr, err)
		require.Nil(t, verifyCodeData)
	})

	t.Run("num allowed equals zero, should return false", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSecureOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedAndIncreaseTrialsCalled: func(key string, _ redis.Mode) (*redis.RateLimiterResult, error) {
				wasCalled = true
				return &redis.RateLimiterResult{Allowed: false}, nil
			},
		}
		totp, _ := secureOtp.NewSecureOtpHandler(args)

		_, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)

		require.True(t, wasCalled)
	})

	t.Run("num allowed equals one, should return true", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSecureOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedAndIncreaseTrialsCalled: func(key string, _ redis.Mode) (*redis.RateLimiterResult, error) {
				wasCalled = true
				return &redis.RateLimiterResult{Allowed: true, Remaining: 1, ResetAfter: time.Duration(10) * time.Second}, nil
			},
		}
		totp, _ := secureOtp.NewSecureOtpHandler(args)

		codeVerifyData, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)

		require.True(t, wasCalled)
		require.Equal(t, 1, codeVerifyData.RemainingTrials)
		require.Equal(t, 10, codeVerifyData.ResetAfter)
	})

	t.Run("should block after max verifications exceeded", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSecureOtpHandler()

		args.RateLimiter = testscommon.NewRateLimiterMock(3, 10)
		totp, _ := secureOtp.NewSecureOtpHandler(args)

		_, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)
		_, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)
		_, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)
		_, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
	})
	t.Run("otp code verify data", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSecureOtpHandler()
		keyData := make(map[string]*redis.RateLimiterResult)
		remainingNormal := 3
		remainingSecurity := 4
		resetAfterNormal := time.Duration(3) * time.Second
		resetAfterSecurity := time.Duration(10) * time.Second

		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedAndIncreaseTrialsCalled: func(key string, mode redis.Mode) (*redis.RateLimiterResult, error) {
				switch mode {
				case redis.NormalMode:
					if _, ok := keyData[key]; !ok {
						keyData[key] = &redis.RateLimiterResult{
							Allowed:    true,
							Remaining:  remainingNormal,
							ResetAfter: resetAfterNormal,
						}
					}
				case redis.SecurityMode:
					if _, ok := keyData[key]; !ok {
						keyData[key] = &redis.RateLimiterResult{
							Allowed:    true,
							Remaining:  remainingSecurity,
							ResetAfter: resetAfterSecurity,
						}
					}
				default:
					return nil, errors.New("unexpected mode")
				}

				keyData[key].Allowed = keyData[key].Remaining > 0
				if keyData[key].Remaining > 0 {
					keyData[key].Remaining--
				}

				return keyData[key], nil
			},
		}

		totp, _ := secureOtp.NewSecureOtpHandler(args)

		expectedResult := &requests.OTPCodeVerifyData{
			RemainingTrials:             2,
			ResetAfter:                  int(math.Round(resetAfterNormal.Seconds())),
			SecurityModeRemainingTrials: 3,
			SecurityModeResetAfter:      int(math.Round(resetAfterSecurity.Seconds())),
		}
		result, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)
		require.Equal(t, expectedResult, result)

		expectedResult.RemainingTrials = 1
		expectedResult.SecurityModeRemainingTrials = 2
		result, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)
		require.Equal(t, expectedResult, result)

		expectedResult.RemainingTrials = 0
		expectedResult.SecurityModeRemainingTrials = 1
		result, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)
		require.Equal(t, expectedResult, result)

		expectedResult.RemainingTrials = 0
		expectedResult.SecurityModeRemainingTrials = 0
		result, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
		require.Equal(t, expectedResult, result)

		expectedResult.RemainingTrials = 2
		expectedResult.SecurityModeRemainingTrials = 0
		ip2 := "127.0.0.2"
		result, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip2)
		require.Nil(t, err)
		require.Equal(t, expectedResult, result)
	})
}

func TestSecureOtpHandler_DecrementSecurityModeFailedTrials(t *testing.T) {
	t.Parallel()

	args := createMockArgsSecureOtpHandler()

	t.Run("on redis limiter error should return err", func(t *testing.T) {
		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			DecrementSecurityFailuresCalled: func(key string) error {
				wasCalled = true
				return expectedErr
			},
		}

		totp, _ := secureOtp.NewSecureOtpHandler(args)

		err := totp.DecrementSecurityModeFailedTrials(account)
		require.Equal(t, expectedErr, err)
		require.True(t, wasCalled)
	})
	t.Run("redis limiter OK", func(t *testing.T) {
		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			DecrementSecurityFailuresCalled: func(key string) error {
				wasCalled = true
				return nil
			},
		}

		totp, _ := secureOtp.NewSecureOtpHandler(args)

		err := totp.DecrementSecurityModeFailedTrials(account)
		require.Nil(t, err)
		require.True(t, wasCalled)
	})
}

func TestSecureOtpHandler_Reset(t *testing.T) {
	t.Parallel()

	t.Run("reset returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSecureOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			ResetCalled: func(key string) error {
				wasCalled = true
				return expectedErr
			},
		}
		totp, _ := secureOtp.NewSecureOtpHandler(args)
		require.NotNil(t, totp)

		totp.Reset(account, ip)

		require.True(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSecureOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			ResetCalled: func(key string) error {
				wasCalled = true
				return nil
			},
		}
		totp, _ := secureOtp.NewSecureOtpHandler(args)
		require.NotNil(t, totp)

		totp.Reset(account, ip)

		require.True(t, wasCalled)
	})
}

func TestSecureOtpHandler_ExtendSecurityMode(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createMockArgsSecureOtpHandler()
	args.RateLimiter = &testscommon.RateLimiterStub{
		ExtendSecurityModeCalled: func(key string) error {
			wasCalled = true
			return nil
		},
	}

	totp, _ := secureOtp.NewSecureOtpHandler(args)
	require.NotNil(t, totp)

	err := totp.ExtendSecurityMode(account)
	require.NoError(t, err)
	require.True(t, wasCalled)
}

func TestSecureOtpHandler_SetSecurityModeNoExpireShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgsSecureOtpHandler()

	wasCalled := false
	args.RateLimiter = &testscommon.RateLimiterStub{
		SetSecurityModeNoExpireCalled: func(key string) error {
			wasCalled = true
			return nil
		},
	}
	totp, _ := secureOtp.NewSecureOtpHandler(args)
	require.NotNil(t, totp)

	err := totp.SetSecurityModeNoExpire(account)
	require.Nil(t, err)

	require.True(t, wasCalled)
}

func TestSecureOtpHandler_SetSecurityModeNoExpireShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgsSecureOtpHandler()
	expectedErr := errors.New("expected error")

	args.RateLimiter = &testscommon.RateLimiterStub{
		SetSecurityModeNoExpireCalled: func(key string) error {
			return expectedErr
		},
	}
	totp, _ := secureOtp.NewSecureOtpHandler(args)
	require.NotNil(t, totp)

	err := totp.SetSecurityModeNoExpire(account)
	require.Equal(t, expectedErr, err)

}

func TestSecureOtpHandler_UnsetSecurityModeNoExpireShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgsSecureOtpHandler()

	wasCalled := false
	args.RateLimiter = &testscommon.RateLimiterStub{
		UnsetSecurityModeNoExpireCalled: func(key string) error {
			wasCalled = true
			return nil
		},
	}
	totp, _ := secureOtp.NewSecureOtpHandler(args)
	require.NotNil(t, totp)

	err := totp.UnsetSecurityModeNoExpire(account)
	require.Nil(t, err)

	require.True(t, wasCalled)
}

func TestSecureOtpHandler_UnsetSecurityModeNoExpireShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgsSecureOtpHandler()
	expectedErr := errors.New("expected error")

	args.RateLimiter = &testscommon.RateLimiterStub{
		UnsetSecurityModeNoExpireCalled: func(key string) error {
			return expectedErr
		},
	}
	totp, _ := secureOtp.NewSecureOtpHandler(args)
	require.NotNil(t, totp)

	err := totp.UnsetSecurityModeNoExpire(account)
	require.Equal(t, expectedErr, err)

}

func TestSecureOtpHandler_Getters(t *testing.T) {
	t.Parallel()

	providedNormalModePeriod := time.Second
	providedNormalModeRate := 1
	providedSecurityModePeriod := time.Minute
	providedSecurityModeRate := 100
	args := createMockArgsSecureOtpHandler()
	args.RateLimiter = &testscommon.RateLimiterStub{
		PeriodCalled: func(mode redis.Mode) time.Duration {
			if mode == redis.NormalMode {
				return providedNormalModePeriod
			}

			return providedSecurityModePeriod
		},
		RateCalled: func(mode redis.Mode) int {
			if mode == redis.NormalMode {
				return providedNormalModeRate
			}

			return providedSecurityModeRate
		},
	}
	totp, _ := secureOtp.NewSecureOtpHandler(args)
	require.NotNil(t, totp)

	require.Equal(t, uint64(providedNormalModePeriod.Seconds()), totp.FreezeBackOffTime())
	require.Equal(t, uint64(providedNormalModeRate), totp.FreezeMaxFailures())
	require.Equal(t, uint64(providedSecurityModePeriod.Seconds()), totp.SecurityModeBackOffTime())
	require.Equal(t, uint64(providedSecurityModeRate), totp.SecurityModeMaxFailures())
}
