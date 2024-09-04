package groups_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	chainApiErrors "github.com/multiversx/mx-chain-go/api/errors"
	chainApiShared "github.com/multiversx/mx-chain-go/api/shared"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/api/groups"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/resolver"
	mockFacade "github.com/multiversx/mx-multi-factor-auth-go-service/testscommon/facade"
)

var (
	expectedError  = errors.New("expected error")
	wrongCodeError = errors.New("wrong code")
	providedAddr   = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
)

func TestNewGuardianGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		gg, err := groups.NewGuardianGroup(nil)

		assert.Nil(t, gg)
		assert.True(t, errors.Is(err, core.ErrNilFacadeHandler))
	})

	t.Run("should work", func(t *testing.T) {
		ng, err := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		assert.NotNil(t, ng)
		assert.Nil(t, err)
	})
}

func TestGuardianGroup_signTransaction(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/sign-transaction", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignTransactionCalled: func(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error) {
				return nil, nil, expectedError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignTransaction{
			Tx: transaction.FrontendTransaction{},
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-transaction", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("facade returns wrong code", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignTransactionCalled: func(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error) {
				return nil, nil, wrongCodeError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignTransaction{
			Tx: transaction.FrontendTransaction{},
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-transaction", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, wrongCodeError.Error()))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestGuardianGroup_signMessage(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/sign-message", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignMessageCalled: func(userIp string, request requests.SignMessage) ([]byte, *requests.OTPCodeVerifyData, error) {
				return nil, nil, expectedError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignMessage{
			Message: "message",
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-message", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("facade returns wrong code", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignMessageCalled: func(userIp string, request requests.SignMessage) ([]byte, *requests.OTPCodeVerifyData, error) {
				return nil, nil, wrongCodeError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignMessage{
			Message: "message",
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-message", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, wrongCodeError.Error()))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("too many failed attempts", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignMessageCalled: func(userIp string, request requests.SignMessage) ([]byte, *requests.OTPCodeVerifyData, error) {
				dataBytes := []byte("signedMsg")
				return dataBytes, &requests.OTPCodeVerifyData{
					RemainingTrials: 0,
					ResetAfter:      123,
				}, core.ErrTooManyFailedAttempts
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignTransaction{
			Tx: transaction.FrontendTransaction{},
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-message", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		resp2 := httptest.NewRecorder()
		ws.ServeHTTP(resp2, req)

		expectedSignMessageResponse := requests.OTPCodeVerifyDataResponse{
			VerifyData: &requests.OTPCodeVerifyData{
				RemainingTrials: 0,
				ResetAfter:      123,
			},
		}

		type DataSignMessageResponse struct {
			Data  requests.OTPCodeVerifyDataResponse `json:"data"`
			Code  string                             `json:"code"`
			Error string                             `json:"error"`
		}

		statusRsp := &DataSignMessageResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Equal(t, expectedSignMessageResponse, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, core.ErrTooManyFailedAttempts.Error()))
		require.Equal(t, http.StatusTooManyRequests, resp.Code)

		statusRsp = &DataSignMessageResponse{}
		loadResponse(resp2.Body, &statusRsp)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedSignMessageResponse := requests.SignMessageResponse{
			Message:   "message",
			Signature: "7369676e65644d657373616765",
		}

		facade := mockFacade.GuardianFacadeStub{
			SignMessageCalled: func(userIp string, request requests.SignMessage) ([]byte, *requests.OTPCodeVerifyData, error) {
				dataBytes := []byte("signedMessage")
				return dataBytes, nil, nil
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignMessage{
			Message: "message",
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-message", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		type DataSignMessageResponse struct {
			Data  requests.SignMessageResponse `json:"data"`
			Code  string                       `json:"code"`
			Error string                       `json:"error"`
		}
		responseData := DataSignMessageResponse{}
		loadResponse(resp.Body, &responseData)

		assert.Equal(t, expectedSignMessageResponse, responseData.Data)
		assert.Equal(t, "", responseData.Error)
		require.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestGuardianGroup_setSecurityModeNoExpire(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/set-security-mode-no-expire", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SetSecurityModeNoExpireCalled: func(userIp string, request requests.SetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error) {
				return nil, expectedError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SetSecurityModeNoExpireMessage{
			UserAddr: providedAddr,
		}
		req, _ := http.NewRequest("POST", "/guardian/set-security-mode-no-expire", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("facade returns wrong code", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SetSecurityModeNoExpireCalled: func(userIp string, request requests.SetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error) {
				return nil, wrongCodeError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SetSecurityModeNoExpireMessage{
			UserAddr: providedAddr,
		}
		req, _ := http.NewRequest("POST", "/guardian/set-security-mode-no-expire", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, wrongCodeError.Error()))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SetSecurityModeNoExpireCalled: func(userIp string, request requests.SetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error) {
				return nil, nil
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SetSecurityModeNoExpireMessage{
			UserAddr: providedAddr,
		}
		req, _ := http.NewRequest("POST", "/guardian/set-security-mode-no-expire", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		type DataSetSecurityModeNoExpire struct {
			Data  string `json:"data"`
			Code  string `json:"code"`
			Error string `json:"error"`
		}
		responseData := DataSetSecurityModeNoExpire{}
		loadResponse(resp.Body, &responseData)

		assert.Equal(t, "", responseData.Data)
		assert.Equal(t, "", responseData.Error)
		require.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestGuardianGroup_unsetSecurityModeNoExpire(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/unset-security-mode-no-expire", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			UnsetSecurityModeNoExpireCalled: func(userIp string, request requests.UnsetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error) {
				return nil, expectedError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.UnsetSecurityModeNoExpireMessage{
			UserAddr: providedAddr,
		}
		req, _ := http.NewRequest("POST", "/guardian/unset-security-mode-no-expire", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("facade returns wrong code", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			UnsetSecurityModeNoExpireCalled: func(userIp string, request requests.UnsetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error) {
				return nil, wrongCodeError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.UnsetSecurityModeNoExpireMessage{
			UserAddr: providedAddr,
		}
		req, _ := http.NewRequest("POST", "/guardian/unset-security-mode-no-expire", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, wrongCodeError.Error()))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			UnsetSecurityModeNoExpireCalled: func(userIp string, request requests.UnsetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error) {
				return nil, nil
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.UnsetSecurityModeNoExpireMessage{
			UserAddr: providedAddr,
		}
		req, _ := http.NewRequest("POST", "/guardian/unset-security-mode-no-expire", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		type DataUnsetSecurityModeNoExpire struct {
			Data  string `json:"data"`
			Code  string `json:"code"`
			Error string `json:"error"`
		}
		responseData := DataUnsetSecurityModeNoExpire{}
		loadResponse(resp.Body, &responseData)

		assert.Equal(t, "", responseData.Data)
		assert.Equal(t, "", responseData.Error)
		require.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestGuardianGroup_signMultipleTransaction(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/sign-multiple-transactions", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignMultipleTransactionsCalled: func(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error) {
				return nil, nil, expectedError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignTransaction{
			Tx: transaction.FrontendTransaction{},
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-multiple-transactions", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("facade returns wrong code", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignMultipleTransactionsCalled: func(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error) {
				return nil, nil, wrongCodeError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignTransaction{
			Tx: transaction.FrontendTransaction{},
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-multiple-transactions", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, wrongCodeError.Error()))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("createSignMultipleTransactionsResponse returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignMultipleTransactionsCalled: func(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error) {
				dummyData, _ := json.Marshal("dummy data")
				return [][]byte{dummyData}, nil, nil
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignTransaction{
			Tx: transaction.FrontendTransaction{},
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-multiple-transactions", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "cannot unmarshal"))
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedUnmarshalledTxs := []transaction.FrontendTransaction{
			{
				Nonce:             1,
				Signature:         "signature",
				GuardianSignature: "guardianSignature",
			},
			{
				Nonce:             2,
				Signature:         "signature",
				GuardianSignature: "guardianSignature",
			},
		}

		expectedSignTransactionResponse := requests.SignMultipleTransactionsResponse{
			Txs: expectedUnmarshalledTxs,
		}

		facade := mockFacade.GuardianFacadeStub{
			SignMultipleTransactionsCalled: func(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error) {
				marshalledTxs := make([][]byte, 0)
				for _, tx := range request.Txs {
					marshalledTx, _ := json.Marshal(tx)
					marshalledTxs = append(marshalledTxs, marshalledTx)
				}
				return marshalledTxs, nil, nil
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		request := requests.SignMultipleTransactions{
			Txs: expectedUnmarshalledTxs,
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-multiple-transactions", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		type SignMultipleTransactionsAPIResponse struct {
			Data  requests.SignMultipleTransactionsResponse `json:"data"`
			Code  string                                    `json:"code"`
			Error string                                    `json:"error"`
		}
		response := SignMultipleTransactionsAPIResponse{}
		loadResponse(resp.Body, &response)

		assert.Equal(t, expectedSignTransactionResponse, response.Data)
		assert.Equal(t, "", response.Error)
		require.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestGuardianGroup_register(t *testing.T) {
	t.Parallel()

	t.Run("empty address", func(t *testing.T) {
		t.Parallel()

		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), "")

		req, _ := http.NewRequest("POST", "/guardian/register", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "bech32"))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/register", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			RegisterUserCalled: func(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error) {
				return &requests.OTP{}, "", expectedError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/register", requestToReader(requests.RegistrationPayload{}))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedData := &requests.RegisterReturnData{
			OTP: &requests.OTP{},
		}
		expectedGenResponse := createExpectedGeneralResponse(expectedData, expectedError.Error())
		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.Equal(t, expectedGenResponse.Error, statusRsp.Error)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedOtpInfo := &requests.OTP{
			Secret: "secret",
		}
		expectedGuardian := "guardian"
		facade := mockFacade.GuardianFacadeStub{
			RegisterUserCalled: func(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error) {
				return expectedOtpInfo, expectedGuardian, nil
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/register", requestToReader(requests.RegistrationPayload{}))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedData := &requests.RegisterReturnData{
			OTP:             expectedOtpInfo,
			GuardianAddress: expectedGuardian,
		}
		expectedErr := ""
		expectedGenResponse := createExpectedGeneralResponse(expectedData, expectedErr)

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.Equal(t, expectedGenResponse.Error, statusRsp.Error)
		require.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestGuardianGroup_verifyCode(t *testing.T) {
	t.Parallel()

	t.Run("empty address", func(t *testing.T) {
		t.Parallel()

		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), "")

		req, _ := http.NewRequest("POST", "/guardian/verify-code", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "bech32"))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/verify-code", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			VerifyCodeCalled: func(userAddress sdkCore.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error) {
				return nil, expectedError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/verify-code", requestToReader(requests.RegistrationPayload{}))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("facade returns wrong code", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			VerifyCodeCalled: func(userAddress sdkCore.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error) {
				return nil, wrongCodeError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/verify-code", requestToReader(requests.RegistrationPayload{}))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedGenResponse := createExpectedGeneralResponse(&requests.OTPCodeVerifyDataResponse{}, "")

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, wrongCodeError.Error()))
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			VerifyCodeCalled: func(userAddress sdkCore.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error) {
				return nil, nil
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("POST", "/guardian/verify-code", requestToReader(requests.RegistrationPayload{}))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		require.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestGuardianGroup_registeredUsers(t *testing.T) {
	t.Parallel()

	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			RegisteredUsersCalled: func() (uint32, error) {
				return 0, expectedError
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("GET", "/guardian/registered-users", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedCount := uint32(150)
		facade := mockFacade.GuardianFacadeStub{
			RegisteredUsersCalled: func() (uint32, error) {
				return expectedCount, nil
			},
		}

		gg, _ := groups.NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

		req, _ := http.NewRequest("GET", "/guardian/registered-users", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedData := &requests.RegisteredUsersResponse{
			Count: expectedCount,
		}
		expectedErr := ""
		expectedGenResponse := createExpectedGeneralResponse(expectedData, expectedErr)

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.Equal(t, expectedGenResponse.Error, statusRsp.Error)
		require.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestGuardianGroup_config(t *testing.T) {
	t.Parallel()

	providedConfig := &core.TcsConfig{
		OTPDelay:         100,
		BackoffWrongCode: 10,
	}

	facade := mockFacade.GuardianFacadeStub{
		TcsConfigCalled: func() *core.TcsConfig {
			return providedConfig
		},
	}

	gg, _ := groups.NewGuardianGroup(&facade)

	ws := startWebServer(gg, "guardian", getServiceRoutesConfig(), providedAddr)

	req, _ := http.NewRequest("GET", "/guardian/config", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := generalResponse{}
	loadResponse(resp.Body, &response)

	expectedData := &requests.ConfigResponse{
		RegistrationDelay: uint32(providedConfig.OTPDelay),
		BackoffWrongCode:  uint32(providedConfig.BackoffWrongCode),
	}
	expectedErr := ""
	expectedGenResponse := createExpectedGeneralResponse(expectedData, expectedErr)

	assert.Equal(t, expectedGenResponse.Data, response.Data)
	assert.Equal(t, expectedGenResponse.Error, response.Error)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestGuardianGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		err := gg.UpdateFacade(nil)
		assert.Equal(t, chainApiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		gg, _ := groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		newFacade := &mockFacade.GuardianFacadeStub{}

		err := gg.UpdateFacade(newFacade)
		assert.Nil(t, err)
	})
}

func TestGuardianGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	gg, _ := groups.NewGuardianGroup(nil)
	assert.True(t, gg.IsInterfaceNil())

	gg, _ = groups.NewGuardianGroup(&mockFacade.GuardianFacadeStub{})
	assert.False(t, gg.IsInterfaceNil())
}

func createExpectedGeneralResponse(data interface{}, expectedErr string) *generalResponse {
	expectedResponse := generalResponse{
		Data:  data,
		Error: expectedErr,
	}
	expectedResponseReader := requestToReader(expectedResponse)
	expectedGenResp := generalResponse{}
	loadResponse(expectedResponseReader, &expectedGenResp)

	return &expectedGenResp
}

func TestHandleHTTPError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err            string
		httpStatusCode int
		returnCode     chainApiShared.ReturnCode
	}{
		{wrongCodeError.Error(), http.StatusBadRequest, chainApiShared.ReturnCodeRequestError},
		{resolver.ErrTooManyTransactionsToSign.Error(), http.StatusBadRequest, chainApiShared.ReturnCodeRequestError},
		{resolver.ErrNoTransactionToSign.Error(), http.StatusBadRequest, chainApiShared.ReturnCodeRequestError},
		{resolver.ErrGuardianMismatch.Error(), http.StatusBadRequest, chainApiShared.ReturnCodeRequestError},
		{resolver.ErrInvalidSender.Error(), http.StatusBadRequest, chainApiShared.ReturnCodeRequestError},
		{resolver.ErrInvalidGuardian.Error(), http.StatusBadRequest, chainApiShared.ReturnCodeRequestError},
		{resolver.ErrGuardianNotUsable.Error(), http.StatusBadRequest, chainApiShared.ReturnCodeRequestError},
		{resolver.ErrGuardianMismatch.Error(), http.StatusBadRequest, chainApiShared.ReturnCodeRequestError},
		{core.ErrTooManyFailedAttempts.Error(), http.StatusTooManyRequests, chainApiShared.ReturnCodeRequestError},
		{handlers.ErrRegistrationFailed.Error(), http.StatusForbidden, chainApiShared.ReturnCodeRequestError},
		{"other internal error", http.StatusInternalServerError, chainApiShared.ReturnCodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.err, func(t *testing.T) {
			httpStatusCode, retCode := groups.HandleHTTPError(tt.err)
			require.Equal(t, tt.httpStatusCode, httpStatusCode)
			require.Equal(t, tt.returnCode, retCode)
		})
	}
}
