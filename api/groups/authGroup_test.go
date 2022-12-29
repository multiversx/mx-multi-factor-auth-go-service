package groups

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	elrondApiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	mockFacade "github.com/ElrondNetwork/multi-factor-auth-go-service/testscommon/facade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedError = errors.New("expected error")

func TestNewNodeGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		ng, err := NewAuthGroup(nil)

		assert.True(t, check.IfNil(ng))
		assert.True(t, errors.Is(err, elrondApiErrors.ErrNilFacadeHandler))
	})
	t.Run("should work", func(t *testing.T) {
		ng, err := NewAuthGroup(&mockFacade.AuthFacadeStub{})

		assert.False(t, check.IfNil(ng))
		assert.Nil(t, err)
	})
}

func TestAuthGroup_signTransaction(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		ag, _ := NewAuthGroup(&mockFacade.AuthFacadeStub{})

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		req, _ := http.NewRequest("POST", "/auth/sign-transaction", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		assert.True(t, strings.Contains(statusRsp.Error, ErrValidation.Error()))
		require.Equal(t, resp.Code, http.StatusInternalServerError)

	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.AuthFacadeStub{
			SignTransactionCalled: func(request requests.SignTransaction) ([]byte, error) {
				return nil, expectedError
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.SignTransaction{
			Tx: data.Transaction{},
		}
		req, _ := http.NewRequest("POST", "/auth/sign-transaction", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		assert.True(t, strings.Contains(statusRsp.Error, ErrValidation.Error()))
		require.Equal(t, resp.Code, http.StatusInternalServerError)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedMarshalledTx := []byte("hash")
		facade := mockFacade.AuthFacadeStub{
			SignTransactionCalled: func(request requests.SignTransaction) ([]byte, error) {
				return expectedMarshalledTx, nil
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.SignTransaction{
			Tx: data.Transaction{},
		}
		req, _ := http.NewRequest("POST", "/auth/sign-transaction", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Equal(t, base64.StdEncoding.EncodeToString(expectedMarshalledTx), statusRsp.Data)
		assert.Equal(t, "", statusRsp.Error)
		require.Equal(t, resp.Code, http.StatusOK)
	})
}

func TestAuthGroup_signMultipleTransaction(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		ag, _ := NewAuthGroup(&mockFacade.AuthFacadeStub{})

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		req, _ := http.NewRequest("POST", "/auth/sign-multiple-transactions", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		assert.True(t, strings.Contains(statusRsp.Error, ErrValidation.Error()))
		require.Equal(t, resp.Code, http.StatusInternalServerError)

	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.AuthFacadeStub{
			SignMultipleTransactionsCalled: func(request requests.SignMultipleTransactions) ([][]byte, error) {
				return nil, expectedError
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.SignTransaction{
			Tx: data.Transaction{},
		}
		req, _ := http.NewRequest("POST", "/auth/sign-multiple-transactions", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		assert.True(t, strings.Contains(statusRsp.Error, ErrValidation.Error()))
		require.Equal(t, resp.Code, http.StatusInternalServerError)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedHashes := [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")}
		facade := mockFacade.AuthFacadeStub{
			SignMultipleTransactionsCalled: func(request requests.SignMultipleTransactions) ([][]byte, error) {
				return expectedHashes, nil
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.SignTransaction{
			Tx: data.Transaction{},
		}
		req, _ := http.NewRequest("POST", "/auth/sign-multiple-transactions", requestToReader(request))
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

func TestAuthGroup_register(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		ag, _ := NewAuthGroup(&mockFacade.AuthFacadeStub{})

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(""))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, "EOF"))
		assert.True(t, strings.Contains(statusRsp.Error, ErrRegister.Error()))
		require.Equal(t, resp.Code, http.StatusInternalServerError)

	})
	t.Run("facade returns error", func(t *testing.T) {
		t.Parallel()

		facade := mockFacade.AuthFacadeStub{
			RegisterUserCalled: func(request requests.RegistrationPayload) ([]byte, string, error) {
				return make([]byte, 0), "", expectedError
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.RegistrationPayload{
			Credentials: "credentials",
		}
		req, _ := http.NewRequest("POST", "/auth/register", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Nil(t, statusRsp.Data)
		assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
		assert.True(t, strings.Contains(statusRsp.Error, ErrRegister.Error()))
		require.Equal(t, resp.Code, http.StatusInternalServerError)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedQr := []byte("qr")
		expectedGuardian := "guardian"
		facade := mockFacade.AuthFacadeStub{
			RegisterUserCalled: func(request requests.RegistrationPayload) ([]byte, string, error) {
				return expectedQr, expectedGuardian, nil
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.RegistrationPayload{
			Credentials: "credentials",
		}
		req, _ := http.NewRequest("POST", "/auth/register", requestToReader(request))
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
		ng, _ := NewAuthGroup(&mockFacade.AuthFacadeStub{})

		err := ng.UpdateFacade(nil)
		assert.Equal(t, elrondApiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		ng, _ := NewAuthGroup(&mockFacade.AuthFacadeStub{})

		newFacade := &mockFacade.AuthFacadeStub{}

		err := ng.UpdateFacade(newFacade)
		assert.Nil(t, err)
		assert.True(t, ng.facade == newFacade) // pointer testing
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
