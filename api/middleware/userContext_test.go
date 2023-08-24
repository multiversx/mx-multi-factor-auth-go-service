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

func createUserContext() *userContext {
	return NewUserContext()
}

func TestUserContextMiddleware(t *testing.T) {
	t.Parallel()

	cfProxy := "178.128.139.205"
	nginxProxy := "178.128.139.204"
	dummyIP := "127.0.0.1:8081"
	providedUserAgent := "Test User Agent"

	t.Run("client connected directly to server, should take from RemoteAddr", func(t *testing.T) {

		providedUserIP := "78.78.78.78:1234"
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
		ws.SetTrustedProxies(nil)

		userContextMiddleware := createUserContext()
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

	t.Run("with proxies, should find in trusted platform header", func(t *testing.T) {
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
		ws.TrustedPlatform = "CF-Connecting-Ip"

		userContextMiddleware := createUserContext()
		require.False(t, check.IfNil(userContextMiddleware))
		ws.Use(userContextMiddleware.MiddlewareHandlerFunc())

		ginAddressRoutes := ws.Group("/guardian")
		ginAddressRoutes.Handle(http.MethodGet, "/test", handlerFunc)

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)

		req.Header.Set("User-Agent", providedUserAgent)
		req.Header.Set("CF-Connecting-Ip", providedUserIP)
		req.Header.Set("X-Real-Ip", providedUserIP)
		req.Header.Set("X-Forwarded-For", xForwardedFor)

		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})

	t.Run("with proxies, if not trusted platform header, should find in x-forwarded-for", func(t *testing.T) {
		providedUserIP := "78.78.78.78"
		xForwardedFor := dummyIP + ", " + providedUserIP + ", " + cfProxy + ", " + nginxProxy

		handlerFunc := func(c *gin.Context) {
			userAgent, exists := c.Get(UserAgentKey)
			assert.True(t, exists, "User agent not found in context")
			assert.Equal(t, providedUserAgent, userAgent)

			userIp, exists := c.Get(UserIpKey)
			assert.True(t, exists, "User IP not found in context")
			assert.Equal(t, providedUserIP, userIp)

			c.Status(http.StatusOK)
		}

		ws := gin.New()
		ws.Use(cors.Default())
		ws.ForwardedByClientIP = true
		ws.SetTrustedProxies([]string{
			"178.128.0.0/16",
		})

		userContextMiddleware := createUserContext()
		require.False(t, check.IfNil(userContextMiddleware))
		ws.Use(userContextMiddleware.MiddlewareHandlerFunc())

		ginAddressRoutes := ws.Group("/guardian")
		ginAddressRoutes.Handle(http.MethodGet, "/test", handlerFunc)

		req, _ := http.NewRequest(http.MethodGet, "/guardian/test", nil)

		req.Header.Set("User-Agent", providedUserAgent)
		req.Header.Set("X-Forwarded-For", xForwardedFor)

		// RemoteAddr set to latest proxy; it expects also port
		req.RemoteAddr = cfProxy + ":8080"

		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)
		assert.Equal(t, resp.Code, http.StatusOK)
	})
}
