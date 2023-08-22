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

	xForwardedForHeader = "x-forwarded-for"
)

// ArgsUserContext defines the arguments needed to create user context
type ArgsUserContext struct {
	UserIPHeaderKeys           []string
	NumProxiesXForwardedHeader int
}

type userContext struct {
	userIpHeaderKeys           []string
	numProxiesXForwardedHeader int
}

// NewUserContext returns a new instance of userContext
func NewUserContext(args ArgsUserContext) *userContext {
	return &userContext{
		userIpHeaderKeys:           args.UserIPHeaderKeys,
		numProxiesXForwardedHeader: args.NumProxiesXForwardedHeader,
	}
}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests
func (uc *userContext) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.Request.Header.Get("user-agent")

		// Be careful when using http headers for getting the real ip since it might
		// depend on multiple factors: intermediate proxies (use information from proxies
		// only if they are trustworthy), custom headers (like for cloudfare, nginx).
		clientIP := uc.getClientIP(c.Request)

		c.Set(UserAgentKey, userAgent)
		c.Set(UserIpKey, clientIP)
		c.Next()
	}
}

// if the server is exposed directly, RemoteAddr should contain the client ip
// if the server is behind trusted proxies, the additional headers can be used (carefully)
func (uc *userContext) getClientIP(r *http.Request) string {
	var xForwardedFor, remoteAddr string

	ip := uc.parseCustomHeaders(r.Header)
	if ip == "" {
		xForwardedFor := r.Header.Get(xForwardedForHeader)
		ip = uc.parseXForwardedFor(xForwardedFor)
	}
	if ip == "" {
		// RemoteAddr contains the last proxy IP or the IP of the client if it is connecting
		// directly to the server, without any proxies in between.
		// It might have the form ip:port so ip has to be extracted
		remoteAddr := r.RemoteAddr
		ip = parseHeader(remoteAddr)
	}

	log.Trace("user context",
		"x-forwarded-for", xForwardedFor,
		"remote-addr", remoteAddr,
		"remoteHost", r.RemoteAddr,
		"host", r.Host,
	)

	return ip
}

func (uc *userContext) parseCustomHeaders(header http.Header) string {
	for _, headerKey := range uc.userIpHeaderKeys {
		ip := parseHeader(header.Get(headerKey))
		if ip != "" {
			return ip
		}
	}

	return ""
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
func (uc *userContext) parseXForwardedFor(header string) string {
	if uc.numProxiesXForwardedHeader == 0 {
		// do not check x-forwarded-for header
		return ""
	}

	addresses := strings.Split(header, ",")
	if len(addresses) == 0 {
		return ""
	}

	var addr string
	if uc.numProxiesXForwardedHeader >= len(addresses) {
		addr = addresses[0]
	} else {
		nAddresses := len(addresses)
		addr = addresses[nAddresses-uc.numProxiesXForwardedHeader]
	}

	trimmedAddr := strings.TrimSpace(addr)

	return parseHeaderEntry(trimmedAddr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (uc *userContext) IsInterfaceNil() bool {
	return uc == nil
}
