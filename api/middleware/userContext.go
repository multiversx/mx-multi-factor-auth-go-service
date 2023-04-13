package middleware

import (
	"github.com/gin-gonic/gin"
)

// UserAgentKey is the key of pair for the user agent stored in the context map
const UserAgentKey = "userAgent"

// UserIpKey is the key of pair for the user ip stored in the context map
const UserIpKey = "userIp"

type userContext struct {
}

// NewUserContext returns a new instance of userContext
func NewUserContext() *userContext {
	return &userContext{}
}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests
func (middleware *userContext) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.Request.Header.Get("user-agent")
		clientIp := c.Request.Header.Get("x-forwarded-for")
		if clientIp == "" {
			clientIp = c.Request.Header.Get("x-real-ip")
		}
		if clientIp == "" {
			clientIp = c.Request.RemoteAddr
		}

		c.Set(UserAgentKey, userAgent)
		c.Set(UserIpKey, clientIp)
		c.Next()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (middleware *userContext) IsInterfaceNil() bool {
	return middleware == nil
}
