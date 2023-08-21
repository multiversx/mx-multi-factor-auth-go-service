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

	t.Run("same ip address", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "192.168.0.1"
		expectedUserIP := "192.168.0.1"
		testUserContextMiddleware(t, providedUserIP, expectedUserIP)
	})

	t.Run("two ipv4 ips", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "192.168.0.1, 192.168.1.1"
		expectedUserIP := "192.168.0.1"
		testUserContextMiddleware(t, providedUserIP, expectedUserIP)
	})

	t.Run("ipv4 with port", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "192.168.0.1:1234, 192.168.1.1"
		expectedUserIP := "192.168.0.1"
		testUserContextMiddleware(t, providedUserIP, expectedUserIP)
	})

	t.Run("localhost ip", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "127.0.0.1, 192.168.1.1, 192.168.0.1"
		expectedUserIP := "127.0.0.1"
		testUserContextMiddleware(t, providedUserIP, expectedUserIP)
	})

	t.Run("ipv6 and ipv4", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "2a02:586:4d31:108d:60bb:e843:aaad:d0f8, 141.101.77.42"
		expectedUserIP := "2a02:586:4d31:108d:60bb:e843:aaad:d0f8"
		testUserContextMiddleware(t, providedUserIP, expectedUserIP)
	})

	t.Run("ipv6 with port", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "[2a02:586:4d31:108d:60bb:e843:aaad:d0f8]:1234, 141.101.77.42"
		expectedUserIP := "2a02:586:4d31:108d:60bb:e843:aaad:d0f8"
		testUserContextMiddleware(t, providedUserIP, expectedUserIP)
	})

	t.Run("with spaces", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "141.101.77.42,   [2a02:586:4d31:108d:60bb:e843:aaad:d0f8]:1234,  192.168.1.1"
		expectedUserIP := "141.101.77.42"
		testUserContextMiddleware(t, providedUserIP, expectedUserIP)
	})
}

func testUserContextMiddleware(t *testing.T, providedUserIP, expectedUserIP string) {
	providedUserAgent := "Test User Agent"

	handlerFunc := func(c *gin.Context) {
		userAgent, exists := c.Get(UserAgentKey)
		assert.True(t, exists, "User agent not found in context")
		assert.Equal(t, providedUserAgent, userAgent)

		userIp, exists := c.Get(UserIpKey)
		assert.True(t, exists, "User IP not found in context")
		assert.Equal(t, expectedUserIP, userIp)

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
		req.Header.Set("User-Agent", providedUserAgent)
		req.Header.Set("X-Forwarded-For", providedUserIP)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
	t.Run("valid value for X-Real-Ip", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)
		req.Header.Set("User-Agent", providedUserAgent)
		req.Header.Set("X-Real-Ip", providedUserIP)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
	t.Run("valud value for RemoteAddr", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)
		req.Header.Set("User-Agent", providedUserAgent)
		req.RemoteAddr = providedUserIP
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
}
