package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/multi-factor-auth-go-service/api/middleware"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("nil status metrics handler", func(t *testing.T) {
		t.Parallel()

		mm, err := middleware.NewMetricsMiddleware(nil)
		require.Nil(t, mm)
		require.Equal(t, core.ErrNilMetricsHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mm, err := middleware.NewMetricsMiddleware(&testscommon.StatusMetricsStub{})
		require.NoError(t, err)
		require.NotNil(t, mm)
	})
}

func TestMetricsMiddleware_MiddlewareHandlerFunc(t *testing.T) {
	t.Parallel()

	type receivedRequestData struct {
		path     string
		duration time.Duration
		status   int
	}

	receivedData := make([]*receivedRequestData, 0)
	mm, err := middleware.NewMetricsMiddleware(&testscommon.StatusMetricsStub{
		AddRequestDataCalled: func(path string, duration time.Duration, status int) {
			receivedData = append(receivedData, &receivedRequestData{
				path:     path,
				duration: duration,
				status:   status,
			})
		},
	})
	require.Nil(t, err)

	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(mm.MiddlewareHandlerFunc())

	handler := func(c *gin.Context) {
		c.JSON(200, "ok")
	}

	ginAddressRoutes := ws.Group("/guardian")
	ginAddressRoutes.Handle(http.MethodGet, "/config", handler)

	resp := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(resp)
	req, _ := http.NewRequestWithContext(context, "GET", "/guardian/config", nil)
	ws.ServeHTTP(resp, req)

	require.Len(t, receivedData, 1)
	require.Equal(t, "/guardian/config", receivedData[0].path)
	require.Equal(t, 200, receivedData[0].status)
}
