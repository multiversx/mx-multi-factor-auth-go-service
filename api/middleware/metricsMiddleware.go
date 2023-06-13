package middleware

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/metrics"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const basePrefix = "/"

type metricsMiddleware struct {
	statusMetricsHandler core.StatusMetricsHandler
	routesConfig         config.ApiRoutesConfig
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// NewMetricsMiddleware returns a new instance of metricsMiddleware
func NewMetricsMiddleware(
	statusMetricsHandler core.StatusMetricsHandler,
	conf config.ApiRoutesConfig,
) (*metricsMiddleware, error) {
	if check.IfNil(statusMetricsHandler) {
		return nil, core.ErrNilMetricsHandler
	}

	mm := &metricsMiddleware{
		statusMetricsHandler: statusMetricsHandler,
		routesConfig:         conf,
	}

	return mm, nil
}

// MiddlewareHandlerFunc handles metrics data in regards to endpoints' durations statistics
func (mm *metricsMiddleware) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		duration := time.Since(t)
		status := c.Writer.Status()

		path := c.FullPath()
		if status == http.StatusOK {
			status = metrics.NonErrorCode
		} else {
			path = mm.getBasePath(c.Request.URL.Path)
		}

		mm.statusMetricsHandler.AddRequestData(path, duration, status)
	}
}

func (mm *metricsMiddleware) getBasePath(path string) string {
	for groupKey, group := range mm.routesConfig.APIPackages {
		routePath := getRouteBasePath(path, groupKey, group.Routes)
		if routePath != "" {
			return routePath
		}

		if strings.HasPrefix(path, basePrefix+groupKey) {
			return basePrefix + groupKey
		}
	}

	return basePrefix
}

func getRouteBasePath(path, groupKey string, routes []config.RouteConfig) string {
	for _, r := range routes {
		if !r.Open {
			continue
		}

		if strings.HasPrefix(path, basePrefix+groupKey+r.Name) {
			return basePrefix + groupKey + r.Name
		}
	}

	return ""
}

// IsInterfaceNil returns true if there is no value under the interface
func (mm *metricsMiddleware) IsInterfaceNil() bool {
	return mm == nil
}
