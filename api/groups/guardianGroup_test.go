package groups

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	mockFacade "github.com/multiversx/multi-factor-auth-go-service/testscommon/facade"
	"github.com/multiversx/mx-chain-core-go/core/check"
	chainApiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedError = errors.New("expected error")

func TestNewNodeGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		gg, err := NewGuardianGroup(nil)

		assert.True(t, check.IfNil(gg))
		assert.True(t, errors.Is(err, chainApiErrors.ErrNilFacadeHandler))
	})
	t.Run("should work", func(t *testing.T) {
		ng, err := NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		assert.False(t, check.IfNil(ng))
		assert.Nil(t, err)
	})
}

func TestGuardianGroup_signTransaction(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig())

		req, _ := http.NewRequest("POST", "/guardian/sign-transaction", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, resp.Code, http.StatusInternalServerError)

	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignTransactionCalled: func(userAddress core.AddressHandler, request requests.SignTransaction) ([]byte, error) {
				return nil, expectedError
			},
		}

		gg, _ := NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig())

		request := requests.SignTransaction{
			Tx: data.Transaction{},
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-transaction", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, resp.Code, http.StatusInternalServerError)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedUnmarshalledTx := data.Transaction{
			Nonce:             1,
			Signature:         "signature",
			GuardianSignature: "guardianSignature",
		}

		expectedSignTransactionResponse := requests.SignTransactionResponse{
			Tx: expectedUnmarshalledTx,
		}

		facade := mockFacade.GuardianFacadeStub{
			SignTransactionCalled: func(userAddress core.AddressHandler, request requests.SignTransaction) ([]byte, error) {
				return json.Marshal(expectedUnmarshalledTx)
			},
		}

		gg, _ := NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig())

		request := requests.SignTransaction{
			Tx: data.Transaction{},
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-transaction", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		type DataSignTransactionResponse struct {
			Data  requests.SignTransactionResponse `json:"data"`
			Code  string                           `json:"code"`
			Error string                           `json:"error"`
		}
		responseData := DataSignTransactionResponse{}
		loadResponse(resp.Body, &responseData)

		assert.Equal(t, expectedSignTransactionResponse, responseData.Data)
		assert.Equal(t, "", responseData.Error)
		require.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestGuardianGroup_signMultipleTransaction(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig())

		req, _ := http.NewRequest("POST", "/guardian/sign-multiple-transactions", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, resp.Code, http.StatusInternalServerError)

	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			SignMultipleTransactionsCalled: func(userAddress core.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
				return nil, expectedError
			},
		}

		gg, _ := NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig())

		request := requests.SignTransaction{
			Tx: data.Transaction{},
		}
		req, _ := http.NewRequest("POST", "/guardian/sign-multiple-transactions", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, resp.Code, http.StatusInternalServerError)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedUnmarshalledTxs := []data.Transaction{
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
			SignMultipleTransactionsCalled: func(userAddress core.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
				marshalledTxs := make([][]byte, 0)
				for _, tx := range request.Txs {
					marshalledTx, _ := json.Marshal(tx)
					marshalledTxs = append( marshalledTxs, marshalledTx)
				}
				return marshalledTxs, nil
			},
		}

		gg, _ := NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig())

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

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		gg, _ := NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig())

		req, _ := http.NewRequest("POST", "/guardian/register", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		require.Equal(t, resp.Code, http.StatusInternalServerError)

	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.GuardianFacadeStub{
			RegisterUserCalled: func(userAddress core.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
				return make([]byte, 0), "", expectedError
			},
		}

		gg, _ := NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig())

		req, _ := http.NewRequest("POST", "/guardian/register", requestToReader(requests.RegistrationPayload{}))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		require.Equal(t, resp.Code, http.StatusInternalServerError)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedQr := []byte("qr")
		expectedGuardian := "guardian"
		facade := mockFacade.GuardianFacadeStub{
			RegisterUserCalled: func(userAddress core.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
				return expectedQr, expectedGuardian, nil
			},
		}

		gg, _ := NewGuardianGroup(&facade)

		ws := startWebServer(gg, "guardian", getServiceRoutesConfig())

		req, _ := http.NewRequest("POST", "/guardian/register", requestToReader(requests.RegistrationPayload{}))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		expectedData := &requests.RegisterReturnData{
			QR:              expectedQr,
			GuardianAddress: expectedGuardian,
		}
		expectedErr := ""
		expectedGenResponse := createExpectedGeneralResponse(expectedData, expectedErr)

		assert.Equal(t, expectedGenResponse.Data, statusRsp.Data)
		assert.Equal(t, expectedGenResponse.Error, statusRsp.Error)
		require.Equal(t, resp.Code, http.StatusOK)
	})
}

func TestNodeGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		gg, _ := NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		err := gg.UpdateFacade(nil)
		assert.Equal(t, chainApiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		gg, _ := NewGuardianGroup(&mockFacade.GuardianFacadeStub{})

		newFacade := &mockFacade.GuardianFacadeStub{}

		err := gg.UpdateFacade(newFacade)
		assert.Nil(t, err)
		assert.True(t, gg.facade == newFacade) // pointer testing
	})
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