package groups

import (
	"encoding/base64"
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

		expectedMarshalledTx := []byte("hash")
		facade := mockFacade.GuardianFacadeStub{
			SignTransactionCalled: func(userAddress core.AddressHandler, request requests.SignTransaction) ([]byte, error) {
				return expectedMarshalledTx, nil
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

		assert.Equal(t, base64.StdEncoding.EncodeToString(expectedMarshalledTx), statusRsp.Data)
		assert.Equal(t, "", statusRsp.Error)
		require.Equal(t, resp.Code, http.StatusOK)
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

		expectedHashes := [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")}
		facade := mockFacade.GuardianFacadeStub{
			SignMultipleTransactionsCalled: func(userAddress core.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
				return expectedHashes, nil
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

		expectedResp := make([]interface{}, len(expectedHashes))
		for idx := range expectedHashes {
			expectedResp[idx] = base64.StdEncoding.EncodeToString(expectedHashes[idx])
		}

		assert.Equal(t, expectedResp, statusRsp.Data)
		assert.Equal(t, "", statusRsp.Error)
		require.Equal(t, resp.Code, http.StatusOK)
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
