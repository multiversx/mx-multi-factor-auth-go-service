package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon/middleware"
	"github.com/stretchr/testify/require"
)

const (
	firstLoggedCode  = "314159"
	secondLoggedCode = "271828"
	setupSecret      = "JBSWY3DPEHPK3PXPSECRET"
)

type syncBuffer struct {
	mut sync.Mutex
	buf bytes.Buffer
}

func (sb *syncBuffer) Write(p []byte) (int, error) {
	sb.mut.Lock()
	defer sb.mut.Unlock()

	return sb.buf.Write(p)
}

func (sb *syncBuffer) String() string {
	sb.mut.Lock()
	defer sb.mut.Unlock()

	return sb.buf.String()
}

func captureLogs(t *testing.T) *syncBuffer {
	buff := &syncBuffer{}
	require.Nil(t, logger.AddLogObserver(buff, &logger.PlainFormatter{}))

	oldPattern := logger.GetLogLevelPattern()
	require.Nil(t, logger.SetLogLevel("*:TRACE"))
	t.Cleanup(func() {
		_ = logger.RemoveLogObserver(buff)
		_ = logger.SetLogLevel(oldPattern)
	})

	return buff
}

func createLoggedEngine(t *testing.T, threshold time.Duration, handler gin.HandlerFunc) *gin.Engine {
	ws := gin.New()
	ws.Use(NewRequestLoggerMiddleware(threshold).MiddlewareHandlerFunc())

	args := createMockArgs()
	args.WhitelistHandler = &middleware.NativeAuthWhitelistHandlerStub{
		IsWhitelistedCalled: func(route string) bool {
			return route != "/guardian/verify-code"
		},
	}
	nativeAuthMiddleware, err := NewNativeAuth(args)
	require.Nil(t, err)
	ws.Use(nativeAuthMiddleware.MiddlewareHandlerFunc())

	guardianGroup := ws.Group("/guardian")
	guardianGroup.POST("/verify-code", handler)
	guardianGroup.POST("/sign-message", handler)
	guardianGroup.POST("/register", handler)

	return ws
}

func sendCodes(ws *gin.Engine, path string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]string{
		"code":       firstLoggedCode,
		"secondCode": secondLoggedCode,
	})
	req, _ := http.NewRequest(http.MethodPost, path+"?code="+firstLoggedCode, bytes.NewReader(body))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	return resp
}

func requireNoSecrets(t *testing.T, logs string) {
	require.False(t, strings.Contains(logs, firstLoggedCode))
	require.False(t, strings.Contains(logs, secondLoggedCode))
	require.False(t, strings.Contains(logs, setupSecret))
}

func TestNewRequestLoggerMiddleware(t *testing.T) {
	t.Parallel()

	rlm := NewRequestLoggerMiddleware(time.Second)
	require.False(t, check.IfNil(rlm))
}

func TestComputeLogTitle(t *testing.T) {
	t.Parallel()

	require.Equal(t, "[too long] api request", computeLogTitle(http.StatusOK))
	require.Equal(t, "[bad request] api request", computeLogTitle(http.StatusBadRequest))
	require.Equal(t, "[internal error] api request", computeLogTitle(http.StatusInternalServerError))
	require.Equal(t, "http code 401 api request", computeLogTitle(http.StatusUnauthorized))
}

func TestRequestLoggerMiddleware_ShouldNotLogBodies(t *testing.T) {
	logs := captureLogs(t)

	t.Run("request rejected by auth before reading the body", func(t *testing.T) {
		ws := createLoggedEngine(t, time.Hour, func(c *gin.Context) {
			require.Fail(t, "should not have been called")
		})

		resp := sendCodes(ws, "/guardian/verify-code")
		require.Equal(t, http.StatusUnauthorized, resp.Code)
		require.Contains(t, logs.String(), "http code 401 api request")
		require.Contains(t, logs.String(), "/guardian/verify-code")
		requireNoSecrets(t, logs.String())
	})
	t.Run("malformed and invalid code requests", func(t *testing.T) {
		statuses := []int{http.StatusBadRequest, http.StatusInternalServerError, http.StatusTooManyRequests}
		for _, status := range statuses {
			ws := createLoggedEngine(t, time.Hour, func(c *gin.Context) {
				c.JSON(status, gin.H{"error": "wrong code " + firstLoggedCode})
			})

			resp := sendCodes(ws, "/guardian/sign-message")
			require.Equal(t, status, resp.Code)
		}
		require.Contains(t, logs.String(), "[bad request] api request")
		require.Contains(t, logs.String(), "[internal error] api request")
		require.Contains(t, logs.String(), "http code 429 api request")
		requireNoSecrets(t, logs.String())
	})
	t.Run("slow successful request with secret in response", func(t *testing.T) {
		ws := createLoggedEngine(t, 0, func(c *gin.Context) {
			body, err := io.ReadAll(c.Request.Body)
			require.Nil(t, err)
			require.Contains(t, string(body), firstLoggedCode)
			c.JSON(http.StatusOK, gin.H{"secret": setupSecret})
		})

		resp := sendCodes(ws, "/guardian/register")
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), setupSecret)
		require.Contains(t, logs.String(), "[too long] api request")
		require.Contains(t, logs.String(), "/guardian/register")
		requireNoSecrets(t, logs.String())
	})
}
