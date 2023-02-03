package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	apiErrors "github.com/multiversx/multi-factor-auth-go-service/api/errors"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-sdk-go/authentication"
	"github.com/multiversx/mx-sdk-go/authentication/native/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func startServerEndpoint(t *testing.T, handler func(c *gin.Context), server authentication.AuthServer, tokenHandler authentication.AuthTokenHandler) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())

	nativeAuthMiddleware, err := NewNativeAuth(server, tokenHandler)
	require.False(t, check.IfNil(nativeAuthMiddleware))
	require.Nil(t, err)
	ws.Use(nativeAuthMiddleware.MiddlewareHandlerFunc())

	ginAddressRoutes := ws.Group("/auth")

	ginAddressRoutes.Handle(http.MethodPost, "/register", handler)
	ginAddressRoutes.Handle(http.MethodGet, "/getRequest", handler)

	return ws
}

func TestNewNativeAuth(t *testing.T) {
	t.Parallel()

	t.Run("nil AuthServer", func(t *testing.T) {
		t.Parallel()

		middleware, err := NewNativeAuth(nil, nil)
		require.Nil(t, middleware)
		require.Equal(t, apiErrors.ErrNilNativeAuthServer, err)
	})
	t.Run("nil TokenHandler", func(t *testing.T) {
		t.Parallel()

		middleware, err := NewNativeAuth(&mock.AuthServerStub{}, nil)
		require.Nil(t, middleware)
		require.Equal(t, authentication.ErrNilTokenHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		middleware, err := NewNativeAuth(&mock.AuthServerStub{}, &mock.AuthTokenHandlerStub{})
		require.False(t, check.IfNil(middleware))
		require.Nil(t, err)
	})
}

func TestNativeAuth_MiddlewareHandlerFunc(t *testing.T) {
	t.Parallel()

	t.Run("no bearer provided", func(t *testing.T) {
		t.Parallel()

		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		ws := startServerEndpoint(t, handlerFunc, &mock.AuthServerStub{}, &mock.AuthTokenHandlerStub{})
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Nil(t, response.Data)
		assert.Equal(t, shared.ReturnCodeRequestError, response.Code)
		assert.Equal(t, ErrMalformedToken.Error(), response.Error)
	})
	t.Run("invalid bearer", func(t *testing.T) {
		t.Parallel()

		server := &mock.AuthServerStub{
			ValidateCalled: func(accessToken authentication.AuthToken) error {
				return expectedErr
			},
		}
		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		ws := startServerEndpoint(t, handlerFunc, server, &mock.AuthTokenHandlerStub{})
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", nil)
		req.Header.Set("Authorization", "Bearer token")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Nil(t, response.Data)
		assert.Equal(t, shared.ReturnCodeRequestError, response.Code)
		assert.Equal(t, expectedErr.Error(), response.Error)
	})
	t.Run("tokenHandler errors should error", func(t *testing.T) {
		t.Parallel()

		server := &mock.AuthServerStub{
			ValidateCalled: func(accessToken authentication.AuthToken) error {
				return expectedErr
			},
		}
		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		tokenHandler := &mock.AuthTokenHandlerStub{
			DecodeCalled: func(accessToken string) (authentication.AuthToken, error) {
				return nil, expectedErr
			},
		}
		ws := startServerEndpoint(t, handlerFunc, server, tokenHandler)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", nil)
		req.Header.Set("Authorization", "Bearer token")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Nil(t, response.Data)
		assert.Equal(t, shared.ReturnCodeRequestError, response.Code)
		assert.Equal(t, expectedErr.Error(), response.Error)
	})
	t.Run("get requests does not requires bearer", func(t *testing.T) {
		t.Parallel()

		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		ws := startServerEndpoint(t, handlerFunc, &mock.AuthServerStub{}, &mock.AuthTokenHandlerStub{})
		req, _ := http.NewRequest(http.MethodGet, "/auth/getRequest", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
	t.Run("valid bearer", func(t *testing.T) {
		t.Parallel()

		server := &mock.AuthServerStub{
			ValidateCalled: func(accessToken authentication.AuthToken) error {
				return nil
			},
		}
		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		token := &mock.AuthTokenStub{
			GetAddressCalled: func() []byte {
				return []byte("addr")
			},
		}
		tokenHandler := &mock.AuthTokenHandlerStub{
			GetUnsignedTokenCalled: func(authToken authentication.AuthToken) []byte {
				return []byte("token")
			},
			DecodeCalled: func(accessToken string) (authentication.AuthToken, error) {
				return token, nil
			},
		}
		ws := startServerEndpoint(t, handlerFunc, server, tokenHandler)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", nil)
		req.Header.Set("Authorization", "Bearer token")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
}
