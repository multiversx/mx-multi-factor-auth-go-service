package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	apiErrors "github.com/multiversx/mx-multi-factor-auth-go-service/api/errors"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon/middleware"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-sdk-go/authentication"
	"github.com/multiversx/mx-sdk-go/authentication/native/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func createMockArgs() ArgNativeAuth {
	return ArgNativeAuth{
		Validator:        &mock.AuthServerStub{},
		TokenHandler:     &mock.AuthTokenHandlerStub{},
		WhitelistHandler: &middleware.NativeAuthWhitelistHandlerStub{},
	}
}

func startServerEndpoint(t *testing.T, handler func(c *gin.Context), args ArgNativeAuth) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())

	nativeAuthMiddleware, err := NewNativeAuth(args)
	require.False(t, check.IfNil(nativeAuthMiddleware))
	require.Nil(t, err)
	ws.Use(nativeAuthMiddleware.MiddlewareHandlerFunc())

	ginAddressRoutes := ws.Group("/guardian")

	ginAddressRoutes.Handle(http.MethodPost, "/register", handler)
	ginAddressRoutes.Handle(http.MethodGet, "/getRequest", handler)

	return ws
}

func TestNewNativeAuth(t *testing.T) {
	t.Parallel()

	t.Run("nil AuthServer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Validator = nil
		nativeAuthInstance, err := NewNativeAuth(args)
		require.Nil(t, nativeAuthInstance)
		require.Equal(t, apiErrors.ErrNilNativeAuthServer, err)
	})
	t.Run("nil TokenHandler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TokenHandler = nil
		nativeAuthInstance, err := NewNativeAuth(args)
		require.Nil(t, nativeAuthInstance)
		require.Equal(t, authentication.ErrNilTokenHandler, err)
	})
	t.Run("nil WhitelistHandler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.WhitelistHandler = nil
		nativeAuthInstance, err := NewNativeAuth(args)
		require.Nil(t, nativeAuthInstance)
		require.Equal(t, apiErrors.ErrNilNativeAuthWhitelistHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		nativeAuthInstance, err := NewNativeAuth(createMockArgs())
		require.False(t, check.IfNil(nativeAuthInstance))
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
		ws := startServerEndpoint(t, handlerFunc, createMockArgs())
		req, _ := http.NewRequest(http.MethodPost, "/guardian/register", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Nil(t, response.Data)
		assert.Equal(t, shared.ReturnCodeRequestError, response.Code)
		assert.Contains(t, response.Error, ErrMalformedToken.Error())
		assert.Contains(t, response.Error, "cannot parse JWT token")
	})
	t.Run("invalid bearer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Validator = &mock.AuthServerStub{
			ValidateCalled: func(accessToken authentication.AuthToken) error {
				return expectedErr
			},
		}
		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		ws := startServerEndpoint(t, handlerFunc, args)
		req, _ := http.NewRequest(http.MethodPost, "/guardian/register", nil)
		req.Header.Set("Authorization", "Bearer token")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Nil(t, response.Data)
		assert.Equal(t, shared.ReturnCodeRequestError, response.Code)
		assert.Contains(t, response.Error, expectedErr.Error())
		assert.Contains(t, response.Error, "JWT token validation failed")
	})
	t.Run("tokenHandler errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Validator = &mock.AuthServerStub{
			ValidateCalled: func(accessToken authentication.AuthToken) error {
				return expectedErr
			},
		}
		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		args.TokenHandler = &mock.AuthTokenHandlerStub{
			DecodeCalled: func(accessToken string) (authentication.AuthToken, error) {
				return nil, expectedErr
			},
		}
		ws := startServerEndpoint(t, handlerFunc, args)
		req, _ := http.NewRequest(http.MethodPost, "/guardian/register", nil)
		req.Header.Set("Authorization", "Bearer token")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Nil(t, response.Data)
		assert.Equal(t, shared.ReturnCodeRequestError, response.Code)
		assert.Contains(t, response.Error, expectedErr.Error())
		assert.Contains(t, response.Error, "cannot decode JWT token")
	})
	t.Run("whitelisted routes do not require bearer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.WhitelistHandler = &middleware.NativeAuthWhitelistHandlerStub{
			IsWhitelistedCalled: func(route string) bool {
				return true
			},
		}
		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		ws := startServerEndpoint(t, handlerFunc, args)
		req, _ := http.NewRequest(http.MethodGet, "/guardian/getRequest", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
	t.Run("valid bearer", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Validator = &mock.AuthServerStub{
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
		args.TokenHandler = &mock.AuthTokenHandlerStub{
			GetUnsignedTokenCalled: func(authToken authentication.AuthToken) []byte {
				return []byte("token")
			},
			DecodeCalled: func(accessToken string) (authentication.AuthToken, error) {
				return token, nil
			},
		}
		ws := startServerEndpoint(t, handlerFunc, args)
		req, _ := http.NewRequest(http.MethodPost, "/guardian/register", nil)
		req.Header.Set("Authorization", "Bearer token")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
}
