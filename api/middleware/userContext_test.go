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

	cfProxy := "12.12.12.12"
	nginxProxy := "13.13.13.13"
	dummyIP := "127.0.0.1:8081"
	providedUserAgent := "Test User Agent"

	t.Run("client connected directly to server, should take from RemoteAddr", func(t *testing.T) {

		providedUserIP := "78.78.78.78:124"
		expectedUserIP := "78.78.78.78"

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

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)
		req.Header.Set("User-Agent", providedUserAgent)
		req.RemoteAddr = providedUserIP
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})

	t.Run("with proxies, should find in cloudfare custom header", func(t *testing.T) {
		providedUserIP := "78.78.78.78"
		expectedUserIP := "78.78.78.78"
		xForwardedFor := dummyIP + ", " + providedUserIP + ", " + cfProxy + ", " + nginxProxy

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

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)

		req.Header.Set("User-Agent", providedUserAgent)
		req.Header.Set("CF-Connecting-Ip", providedUserIP)
		req.Header.Set("X-Real-Ip", providedUserIP)
		req.Header.Set("X-Forwarded-For", xForwardedFor)

		// RemoteAddr set to latest proxy
		req.RemoteAddr = cfProxy

		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})

	t.Run("with proxies, if not in cloudfare header, should find in nginx custom header", func(t *testing.T) {
		providedUserIP := "78.78.78.78"
		expectedUserIP := "78.78.78.78"
		xForwardedFor := dummyIP + ", " + providedUserIP + ", " + cfProxy + ", " + nginxProxy

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

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)

		req.Header.Set("User-Agent", providedUserAgent)
		req.Header.Set("X-Real-Ip", providedUserIP)
		req.Header.Set("X-Forwarded-For", xForwardedFor)

		// RemoteAddr set to latest proxy
		req.RemoteAddr = cfProxy

		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})

	t.Run("with proxies, if not custom header, should find in x-forwarded-for", func(t *testing.T) {
		providedUserIP := "78.78.78.78"
		xForwardedFor := dummyIP + ", " + providedUserIP + ", " + cfProxy + ", " + nginxProxy

		handlerFunc := func(c *gin.Context) {
			userAgent, exists := c.Get(UserAgentKey)
			assert.True(t, exists, "User agent not found in context")
			assert.Equal(t, providedUserAgent, userAgent)

			userIp, exists := c.Get(UserIpKey)
			assert.True(t, exists, "User IP not found in context")
			assert.Equal(t, nginxProxy, userIp)

			c.Status(http.StatusOK)
		}

		ws := gin.New()
		ws.Use(cors.Default())

		userContextMiddleware := NewUserContext()
		require.False(t, check.IfNil(userContextMiddleware))
		ws.Use(userContextMiddleware.MiddlewareHandlerFunc())

		ginAddressRoutes := ws.Group("/guardian")
		ginAddressRoutes.Handle(http.MethodGet, "/test", handlerFunc)

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)

		req.Header.Set("User-Agent", providedUserAgent)
		req.Header.Set("X-Forwarded-For", xForwardedFor)

		// RemoteAddr set to latest proxy
		req.RemoteAddr = cfProxy

		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
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

func TestParseHeader(t *testing.T) {
	t.Parallel()

	t.Run("private subnet ip address", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "192.168.0.1"
		expectedUserIP := "192.168.0.1"
		require.Equal(t, expectedUserIP, parseHeader(providedUserIP))
	})

	t.Run("with port", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "78.78.78.78:1234"
		expectedUserIP := "78.78.78.78"
		require.Equal(t, expectedUserIP, parseHeader(providedUserIP))
	})

	// this should not usually happen, RemoteAddr field should contain only one entry
	t.Run("multiple ips, ipv6 and ipv4", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "[2a02:586:4d31:108d:60bb:e843:aaad:d0f8]:1234, 141.101.77.42"
		expectedUserIP := "2a02:586:4d31:108d:60bb:e843:aaad:d0f8"
		require.Equal(t, expectedUserIP, parseHeader(providedUserIP))
	})
}

func TestParseXForwardedFor(t *testing.T) {
	t.Parallel()

	t.Run("same ip address", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "192.168.0.1"
		expectedUserIP := "192.168.0.1"
		require.Equal(t, expectedUserIP, parseXForwardedFor(providedUserIP))
	})

	t.Run("should work with loopback address", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "127.0.0.1"
		expectedUserIP := "127.0.0.1"
		require.Equal(t, expectedUserIP, parseXForwardedFor(providedUserIP))
	})

	t.Run("two ipv4 ips", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "192.168.0.1, 192.168.1.1"
		expectedUserIP := "192.168.1.1"
		require.Equal(t, expectedUserIP, parseXForwardedFor(providedUserIP))
	})

	t.Run("ipv4 with port", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "192.168.0.1:1234"
		expectedUserIP := "192.168.0.1"
		require.Equal(t, expectedUserIP, parseXForwardedFor(providedUserIP))
	})

	t.Run("localhost ip", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "127.0.0.1, 192.168.1.1, 192.168.0.1"
		expectedUserIP := "192.168.0.1"
		require.Equal(t, expectedUserIP, parseXForwardedFor(providedUserIP))
	})

	t.Run("ipv6 and ipv4", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "2a02:586:4d31:108d:60bb:e843:aaad:d0f8"
		expectedUserIP := "2a02:586:4d31:108d:60bb:e843:aaad:d0f8"
		require.Equal(t, expectedUserIP, parseXForwardedFor(providedUserIP))
	})

	t.Run("ipv6 with port", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "[2a02:586:4d31:108d:60bb:e843:aaad:d0f8]:1234, 141.101.77.42"
		expectedUserIP := "141.101.77.42"
		require.Equal(t, expectedUserIP, parseXForwardedFor(providedUserIP))
	})

	t.Run("with spaces", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "141.101.77.42,   [2a02:586:4d31:108d:60bb:e843:aaad:d0f8]:1234,  192.168.1.1"
		expectedUserIP := "192.168.1.1"
		require.Equal(t, expectedUserIP, parseXForwardedFor(providedUserIP))
	})

	t.Run("return empty string if not set correctly", func(t *testing.T) {
		t.Parallel()

		providedUserIP := "wrong test string"
		expectedUserIP := ""
		require.Equal(t, expectedUserIP, parseXForwardedFor(providedUserIP))
	})
}
