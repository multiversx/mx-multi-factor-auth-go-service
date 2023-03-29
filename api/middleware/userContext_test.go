package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserContextMiddleware(t *testing.T) {

	t.Parallel()

	handlerFunc := func(c *gin.Context) {
		userAgent, exists := c.Get(UserAgentKey)
		assert.True(t, exists, "User agent not found in context")
		assert.NotEmpty(t, userAgent, "User agent is empty")

		userIp, exists := c.Get(UserIpKey)
		assert.True(t, exists, "User IP not found in context")
		assert.NotEmpty(t, userIp, "User IP is empty")

		c.Status(http.StatusOK)
	}
	ws := gin.New()
	ws.Use(cors.Default())

	userContextMiddleware := NewUserContext()
	require.False(t, check.IfNil(userContextMiddleware))
	ws.Use(userContextMiddleware.MiddlewareHandlerFunc())

	ginAddressRoutes := ws.Group("/guardian")
	ginAddressRoutes.Handle(http.MethodGet, "/test", handlerFunc)

	t.Run("valid value for X-Forwarded-For", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)
		req.Header.Set("User-Agent", "Test User Agent")
		req.Header.Set("X-Forwarded-For", "192.168.0.1")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
	t.Run("valid value for X-Real-Ip", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)
		req.Header.Set("User-Agent", "Test User Agent")
		req.Header.Set("X-Real-Ip", "192.168.0.1")
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
	t.Run("valud value for RemoteAddr", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)
		req.Header.Set("User-Agent", "Test User Agent")
		req.RemoteAddr = "192.168.0.1"
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
}
