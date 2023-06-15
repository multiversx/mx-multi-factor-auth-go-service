package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/multi-factor-auth-go-service/api/middleware"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/metrics"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
)

func getDefaultRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"status": {
				Routes: []config.RouteConfig{
					{Name: "/metrics", Open: true},
					{Name: "/prometheus-metrics", Open: true},
				},
			},
			"guardian": {
				Routes: []config.RouteConfig{
					{Name: "/register", Open: true},
					{Name: "/verify-code", Open: false},
				},
			},
		},
	}
}

func TestNewMetricsMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("nil status metrics handler", func(t *testing.T) {
		t.Parallel()

		mm, err := middleware.NewMetricsMiddleware(nil, getDefaultRoutesConfig())
		require.Nil(t, mm)
		require.Equal(t, core.ErrNilMetricsHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mm, err := middleware.NewMetricsMiddleware(&testscommon.StatusMetricsStub{}, getDefaultRoutesConfig())
		require.NoError(t, err)
		require.NotNil(t, mm)
	})
}

func TestMetricsMiddleware_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	mm, _ := middleware.NewMetricsMiddleware(nil, getDefaultRoutesConfig())
	require.True(t, mm.IsInterfaceNil())

	mm, _ = middleware.NewMetricsMiddleware(&testscommon.StatusMetricsStub{}, getDefaultRoutesConfig())
	require.False(t, mm.IsInterfaceNil())
}

func TestGetValidPath(t *testing.T) {
	t.Parallel()

	mm, _ := middleware.NewMetricsMiddleware(&testscommon.StatusMetricsStub{}, getDefaultRoutesConfig())

	require.Equal(t, "/", mm.GetBasePath("/asdasdsa?a=10"))
	require.Equal(t, "/", mm.GetBasePath("/wrongpath/metrics/asdasdsa"))

	t.Run("status routes", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "/status", mm.GetBasePath("/status/prom"))
		require.Equal(t, "/status/prometheus-metrics", mm.GetBasePath("/status/prometheus-metrics"))
		require.Equal(t, "/status/metrics", mm.GetBasePath("/status/metrics/asdasdsa"))
		require.Equal(t, "/status/metrics", mm.GetBasePath("/status/metrics/asdasdsa?a=10"))
		require.Equal(t, "/status", mm.GetBasePath("/status/guardian"))
	})

	t.Run("guardian routes", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "/guardian/register", mm.GetBasePath("/guardian/register"))
		require.Equal(t, "/guardian/register", mm.GetBasePath("/guardian/register/asdsadas"))
		require.Equal(t, "/guardian", mm.GetBasePath("/guardian/reg"))
		require.Equal(t, "/guardian", mm.GetBasePath("/guardian/verify-code"))
		require.Equal(t, "/guardian", mm.GetBasePath("/guardian/verify-code/asdsadas"))
		require.Equal(t, "/guardian", mm.GetBasePath("/guardian/status"))
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
	}, getDefaultRoutesConfig())
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
	require.Equal(t, metrics.NonErrorCode, receivedData[0].status)
}
