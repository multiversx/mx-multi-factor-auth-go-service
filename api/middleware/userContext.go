package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// UserAgentKey is the key of pair for the user agent stored in the context map
	UserAgentKey = "userAgent"

	// UserIpKey is the key of pair for the user ip stored in the context map
	UserIpKey = "userIp"

	xForwardedForHeader  = "x-forwarded-for"
	xRealIPHeader        = "x-real-ip"
	cfConnectingIPHeader = "cf-connecting-ip"
)

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

		// Be careful when using http headers for getting the real ip since it might
		// depend on multiple factors: intermediate proxies (use information from proxies
		// only if they are trustworthy), custom headers (like for cloudfare, nginx).
		clientIP := getClientIP(c.Request)

		c.Set(UserAgentKey, userAgent)
		c.Set(UserIpKey, clientIP)
		c.Next()
	}
}

// if the server is exposed directly, RemoteAddr should contain the client ip
// if the server is behind trusted proxies, the additional headers can be used (carefully)
func getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get(xForwardedForHeader)
	xRealIP := r.Header.Get(xRealIPHeader)
	cfConnectingIP := r.Header.Get(cfConnectingIPHeader)
	remoteAddr := r.RemoteAddr

	// TODO: change log level to trace
	log.Debug("user context",
		"x-forwarded-for", xForwardedFor,
		"x-real-ip", xRealIP,
		"remote-addr", remoteAddr,
	)

	ip := parseHeader(cfConnectingIP)
	if ip == "" {
		ip = parseHeader(xRealIP)
	}
	if ip == "" {
		ip = parseXForwardedFor(xForwardedFor)
	}
	if ip == "" {
		// RemoteAddr contains the last proxy IP or the IP of the client if it is connecting
		// directly to the server, without any proxies in between.
		// It might have the form ip:port so ip has to be extracted
		ip = parseHeader(remoteAddr)
	}

	return ip
}

func parseHeader(header string) string {
	addresses := strings.Split(header, ",")

	addr := strings.TrimSpace(addresses[0])

	return parseHeaderEntry(addr)
}

func parseHeaderEntry(header string) string {
	ip, _, err := net.SplitHostPort(header)
	if err == nil {
		return ip
	}

	ipAddr := net.ParseIP(header)
	if ipAddr.IsGlobalUnicast() {
		return ipAddr.String()
	}
	if ipAddr.IsLoopback() {
		return ipAddr.String()
	}

	return ""
}

// when parsing x-forwarded-for header use the rightmost IP from the list, because that's
// the most trustworthy.
// the leftmost IP is the closest to the client, but can be easily spoofed
func parseXForwardedFor(header string) string {
	addresses := strings.Split(header, ",")

	for i := len(addresses) - 1; i >= 0; i-- {
		ip := strings.TrimSpace(addresses[i])

		addr := parseHeaderEntry(ip)
		if addr == "" {
			continue
		}

		return addr
	}

	return ""
}

// IsInterfaceNil returns true if there is no value under the interface
func (middleware *userContext) IsInterfaceNil() bool {
	return middleware == nil
}
