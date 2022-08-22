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
	mockFacade "github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon/facade"
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

func TestAuthGroup_sendTransaction(t *testing.T) {
	t.Parallel()

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		ag, _ := NewAuthGroup(&mockFacade.FacadeStub{})

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		req, _ := http.NewRequest("POST", "/auth/sendTransaction", strings.NewReader(""))
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
			ValidateCalled: func(request requests.SendTransaction) (string, error) {
				return "", expectedError
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.SendTransaction{
			Account: "acc1",
			Codes:   make([]requests.Code, 0),
			Tx:      data.Transaction{},
		}
		req, _ := http.NewRequest("POST", "/auth/sendTransaction", requestToReader(request))
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

		expectedHash := "hash"
		facade := mockFacade.FacadeStub{
			ValidateCalled: func(request requests.SendTransaction) (string, error) {
				return expectedHash, nil
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.SendTransaction{
			Account: "acc1",
			Codes:   make([]requests.Code, 0),
			Tx:      data.Transaction{},
		}
		req, _ := http.NewRequest("POST", "/auth/sendTransaction", requestToReader(request))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		statusRsp := generalResponse{}
		loadResponse(resp.Body, &statusRsp)

		assert.Equal(t, expectedHash, statusRsp.Data)
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
			RegisterUserCalled: func(request requests.Register) ([]byte, error) {
				return make([]byte, 0), expectedError
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.Register{
			Account:  "addr0",
			Provider: "provider",
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
			RegisterUserCalled: func(request requests.Register) ([]byte, error) {
				return expectedQr, nil
			},
		}

		ag, _ := NewAuthGroup(&facade)

		ws := startWebServer(ag, "auth", getServiceRoutesConfig())

		request := requests.Register{
			Account:  "addr0",
			Provider: "provider",
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
