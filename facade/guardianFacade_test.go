package facade

import (
	"errors"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArguments() ArgsGuardianFacade {
	return ArgsGuardianFacade{
		ServiceResolver:      &testscommon.ServiceResolverStub{},
		StatusMetricsHandler: &testscommon.StatusMetricsStub{},
	}
}

func TestNewGuardianFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil service resolver should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ServiceResolver = nil

		facadeInstance, err := NewGuardianFacade(args)
		assert.Nil(t, facadeInstance)
		assert.True(t, errors.Is(err, ErrNilServiceResolver))
	})

	t.Run("nil metrics handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.StatusMetricsHandler = nil

		facadeInstance, err := NewGuardianFacade(args)
		assert.Nil(t, facadeInstance)
		assert.True(t, errors.Is(err, core.ErrNilMetricsHandler))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewGuardianFacade(createMockArguments())
		assert.NotNil(t, facadeInstance)
		assert.Nil(t, err)
	})
}

func TestGuardianFacade_Getters(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	args := createMockArguments()
	expectedGuardian := "expected guardian"
	wasVerifyCodeCalled := false
	providedVerifyCodeReq := requests.VerificationPayload{
		Code:     "VerifyCode code",
		Guardian: "VerifyCode guardian",
	}
	providedUserAddress, _ := data.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
	expectedOtpInfo := &requests.OTP{
		QR:     []byte("expected qr"),
		Secret: "secret",
	}
	wasRegisterUserCalled := false
	otpDelay := uint64(50)
	backoffWrongCode := uint64(100)

	providedIp := "provided ip"
	providedSignTxReq := requests.SignTransaction{
		Code: "123456",
		Tx:   transaction.FrontendTransaction{},
	}
	expectedSignTxResponse := []byte("expected sign tx")
	wasSignTransactionCalled := false

	providedSignMultipleTxsReq := requests.SignMultipleTransactions{
		Code: "123456",
		Txs:  []transaction.FrontendTransaction{},
	}
	expectedSignMultipleTxsResponse := [][]byte{[]byte("expected tx 1 signed"), []byte("expected tx 2 signed")}
	wasSignMultipleTransactionCalled := false

	providedCount := uint32(100)
	wasRegisteredUsersCalled := false

	args.ServiceResolver = &testscommon.ServiceResolverStub{
		VerifyCodeCalled: func(userAddress sdkCore.AddressHandler, userIp string, request requests.VerificationPayload) error {
			assert.Equal(t, providedVerifyCodeReq, request)
			wasVerifyCodeCalled = true
			return nil
		},
		RegisterUserCalled: func(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error) {
			assert.Equal(t, providedUserAddress, userAddress)
			wasRegisterUserCalled = true
			return expectedOtpInfo, expectedGuardian, nil
		},
		TcsConfigCalled: func() *core.TcsConfig {
			return &core.TcsConfig{
				OTPDelay:         otpDelay,
				BackoffWrongCode: backoffWrongCode,
			}
		},
		SignTransactionCalled: func(userIp string, request requests.SignTransaction) ([]byte, error) {
			assert.Equal(t, providedIp, userIp)
			assert.Equal(t, providedSignTxReq, request)
			wasSignTransactionCalled = true
			return expectedSignTxResponse, nil
		},
		SignMultipleTransactionsCalled: func(userIp string, request requests.SignMultipleTransactions) ([][]byte, error) {
			assert.Equal(t, providedIp, userIp)
			assert.Equal(t, providedSignMultipleTxsReq, request)
			wasSignMultipleTransactionCalled = true
			return expectedSignMultipleTxsResponse, nil
		},
		RegisteredUsersCalled: func() (uint32, error) {
			wasRegisteredUsersCalled = true
			return providedCount, nil
		},
	}

	wasGetAllCalled := false
	expectedEndpointMetricsResp := map[string]*requests.EndpointMetricsResponse{
		"m1": {},
	}

	wasGetMetricsForPrometheusCalled := false
	expectedPrometheusMetrics := "expected metrics"
	args.StatusMetricsHandler = &testscommon.StatusMetricsStub{
		GetAllCalled: func() map[string]*requests.EndpointMetricsResponse {
			wasGetAllCalled = true
			return expectedEndpointMetricsResp
		},
		GetMetricsForPrometheusCalled: func() string {
			wasGetMetricsForPrometheusCalled = true
			return expectedPrometheusMetrics
		},
	}
	facadeInstance, _ := NewGuardianFacade(args)

	assert.Nil(t, facadeInstance.VerifyCode(providedUserAddress, "userIp", providedVerifyCodeReq))
	assert.True(t, wasVerifyCodeCalled)

	otpInfo, guardian, err := facadeInstance.RegisterUser(providedUserAddress, requests.RegistrationPayload{})
	assert.Nil(t, err)
	assert.Equal(t, expectedOtpInfo, otpInfo)
	assert.Equal(t, expectedGuardian, guardian)
	assert.True(t, wasRegisterUserCalled)

	tcsConfig := facadeInstance.TcsConfig()
	require.NotNil(t, tcsConfig)
	require.Equal(t, otpDelay, tcsConfig.OTPDelay)
	require.Equal(t, backoffWrongCode, tcsConfig.BackoffWrongCode)

	signedTx, err := facadeInstance.SignTransaction(providedIp, providedSignTxReq)
	assert.Nil(t, err)
	assert.Equal(t, expectedSignTxResponse, signedTx)
	assert.True(t, wasSignTransactionCalled)

	signedTxs, err := facadeInstance.SignMultipleTransactions(providedIp, providedSignMultipleTxsReq)
	assert.Nil(t, err)
	assert.Equal(t, expectedSignMultipleTxsResponse, signedTxs)
	assert.True(t, wasSignMultipleTransactionCalled)

	count, err := facadeInstance.RegisteredUsers()
	assert.Nil(t, err)
	assert.Equal(t, providedCount, count)
	assert.True(t, wasRegisteredUsersCalled)

	metrics := facadeInstance.GetMetrics()
	assert.Equal(t, expectedEndpointMetricsResp, metrics)
	assert.True(t, wasGetAllCalled)

	prometheusMetrics := facadeInstance.GetMetricsForPrometheus()
	assert.Equal(t, expectedPrometheusMetrics, prometheusMetrics)
	assert.True(t, wasGetMetricsForPrometheusCalled)
}

func TestGuardianFacade_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var facadeInstance *guardianFacade
	assert.True(t, facadeInstance.IsInterfaceNil())

	facadeInstance, _ = NewGuardianFacade(createMockArguments())
	assert.False(t, facadeInstance.IsInterfaceNil())
}
