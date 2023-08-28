package middleware

import (
	"github.com/gin-gonic/gin"
)

const (
	// UserAgentKey is the key of pair for the user agent stored in the context map
	UserAgentKey = "userAgent"

	// UserIpKey is the key of pair for the user ip stored in the context map
	UserIpKey = "userIp"
)

type userContext struct {
}

// NewUserContext returns a new instance of userContext
func NewUserContext() *userContext {
	return &userContext{}
}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests
func (uc *userContext) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.Request.Header.Get("user-agent")

		// Be careful when using http headers for getting the real ip since it might
		// depend on multiple factors: intermediate proxies (use information from proxies
		// only if they are trustworthy), custom headers (like for cloudflare, nginx).
		clientIP := c.ClientIP()

		c.Set(UserAgentKey, userAgent)
		c.Set(UserIpKey, clientIP)
		c.Next()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (uc *userContext) IsInterfaceNil() bool {
	return uc == nil
}
