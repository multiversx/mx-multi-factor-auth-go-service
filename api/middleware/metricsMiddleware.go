package middleware

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type metricsMiddleware struct {
	statusMetricsHandler core.StatusMetricsHandler
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// NewMetricsMiddleware returns a new instance of metricsMiddleware
func NewMetricsMiddleware(statusMetricsHandler core.StatusMetricsHandler) (*metricsMiddleware, error) {
	if check.IfNil(statusMetricsHandler) {
		return nil, core.ErrNilMetricsHandler
	}

	mm := &metricsMiddleware{
		statusMetricsHandler: statusMetricsHandler,
	}

	return mm, nil
}

// MiddlewareHandlerFunc handlers metrics data in regards to endpoints' durations statistics
func (mm *metricsMiddleware) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		duration := time.Since(t)
		status := c.Writer.Status()

		mm.statusMetricsHandler.AddRequestData(c.FullPath(), duration, status)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (mm *metricsMiddleware) IsInterfaceNil() bool {
	return mm == nil
}
