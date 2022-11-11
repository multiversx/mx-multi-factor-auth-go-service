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
		ng, err := NewAuthGroup(&mockFacade.FacadeStub{})

		assert.False(t, check.IfNil(ng))
		assert.Nil(t, err)
	})
}

func TestAuthGroup_signTransaction(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		ag, _ := NewAuthGroup(&mockFacade.FacadeStub{})

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

		facade := mockFacade.FacadeStub{
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
		facade := mockFacade.FacadeStub{
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

		ag, _ := NewAuthGroup(&mockFacade.FacadeStub{})

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

		facade := mockFacade.FacadeStub{
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
		facade := mockFacade.FacadeStub{
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

		ag, _ := NewAuthGroup(&mockFacade.FacadeStub{})

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

		facade := mockFacade.FacadeStub{
			RegisterUserCalled: func(request requests.RegistrationPayload) ([]byte, error) {
				return make([]byte, 0), expectedError
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.RegistrationPayload{
			Credentials: "credentials",
			Guardian:    "guardian",
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
		facade := mockFacade.FacadeStub{
			RegisterUserCalled: func(request requests.RegistrationPayload) ([]byte, error) {
				return expectedQr, nil
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.RegistrationPayload{
			Credentials: "credentials",
			Guardian:    "guardian",
		}
		req, _ := http.NewRequest("POST", "/auth/register", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Equal(t, base64.StdEncoding.EncodeToString(expectedQr), statusRsp.Data)
		assert.Equal(t, "", statusRsp.Error)
		require.Equal(t, resp.Code, http.StatusOK)
	})
}

func TestAuthGroup_getGuardianAddress(t *testing.T) {
	t.Parallel()

	expectedAddress := "address"
	facade := mockFacade.FacadeStub{
		GetGuardianAddressCalled: func(request requests.GetGuardianAddress) (string, error) {
			return expectedAddress, nil
		},
	}

	ag, _ := NewAuthGroup(&facade)

	ws := startWebServer(ag, "auth", getServiceRoutesConfig())

	request := requests.GetGuardianAddress{
		Credentials: "credentials",
	}

	req, _ := http.NewRequest("POST", "/auth/generate-guardian", requestToReader(request))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	addrResp := generalResponse{}
	loadResponse(resp.Body, &addrResp)
	assert.Equal(t, expectedAddress, addrResp.Data)
	assert.Equal(t, "", addrResp.Error)
	require.Equal(t, resp.Code, http.StatusOK)
}

func TestNodeGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		ng, _ := NewAuthGroup(&mockFacade.FacadeStub{})

		err := ng.UpdateFacade(nil)
		assert.Equal(t, elrondApiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		ng, _ := NewAuthGroup(&mockFacade.FacadeStub{})

		newFacade := &mockFacade.FacadeStub{}

		err := ng.UpdateFacade(newFacade)
		assert.Nil(t, err)
		assert.True(t, ng.facade == newFacade) // pointer testing
	})
}
