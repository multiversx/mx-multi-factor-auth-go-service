package middleware

import (
	"net"
	"strings"

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

		clientIp := parseIPHeader(c.Request.Header.Get("x-forwarded-for"))
		if clientIp == "" {
			clientIp = parseIPHeader(c.Request.Header.Get("x-real-ip"))
		}
		if clientIp == "" {
			clientIp = parseIPHeader(c.Request.RemoteAddr)
		}

		c.Set(UserAgentKey, userAgent)
		c.Set(UserIpKey, clientIp)
		c.Next()
	}
}

func parseIPHeader(header string) string {
	addresses := strings.Split(header, ",")

	addr := strings.TrimSpace(addresses[0])

	ip, _, err := net.SplitHostPort(addr)
	if err == nil {
		return ip
	}

	realIP := net.ParseIP(addr)
	if !realIP.IsGlobalUnicast() {
		if realIP.IsLoopback() {
			return realIP.String()
		}

		return ""
	}

	return realIP.String()
}

// IsInterfaceNil returns true if there is no value under the interface
func (middleware *userContext) IsInterfaceNil() bool {
	return middleware == nil
}
