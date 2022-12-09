package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	elrondApiShared "github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/authentication"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/authentication/native/mock"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func startServerEndpoint(t *testing.T, handler func(c *gin.Context), server authentication.AuthServer) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())

	nativeAuthMiddleware := NewNativeAuth(server)
	require.False(t, nativeAuthMiddleware.IsInterfaceNil())
	ws.Use(nativeAuthMiddleware.MiddlewareHandlerFunc())

	ginAddressRoutes := ws.Group("/auth")

	ginAddressRoutes.Handle(http.MethodPost, "/register", handler)

	return ws
}

func TestNewNodeGroup(t *testing.T) {
	t.Parallel()

	t.Run("no bearer provided", func(t *testing.T) {
		t.Parallel()

		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		ws := startServerEndpoint(t, handlerFunc, &mock.AuthServerStub{})
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response elrondApiShared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Nil(t, response.Data)
		assert.Equal(t, elrondApiShared.ReturnCodeRequestError, response.Code)
		assert.Equal(t, ErrMalformedToken.Error(), response.Error)
	})
	t.Run("invalid bearer", func(t *testing.T) {
		t.Parallel()

		server := &mock.AuthServerStub{
			ValidateCalled: func(accessToken string) (string, error) {
				return "", expectedErr
			},
		}
		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		ws := startServerEndpoint(t, handlerFunc, server)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", nil)
		req.Header.Set("Authorization", "Bearer token")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response elrondApiShared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Nil(t, response.Data)
		assert.Equal(t, elrondApiShared.ReturnCodeRequestError, response.Code)
		assert.Equal(t, expectedErr.Error(), response.Error)
	})
	t.Run("valid bearer", func(t *testing.T) {
		t.Parallel()

		server := &mock.AuthServerStub{
			ValidateCalled: func(accessToken string) (string, error) {
				return "erd123", nil
			},
		}
		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		ws := startServerEndpoint(t, handlerFunc, server)
		req, _ := http.NewRequest(http.MethodPost, "/auth/register", nil)
		req.Header.Set("Authorization", "Bearer token")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response elrondApiShared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
}
