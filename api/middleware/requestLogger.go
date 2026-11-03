package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	prefixDurationTooLong = "[too long]"
	prefixBadRequest      = "[bad request]"
	prefixInternalError   = "[internal error]"
)

type requestLoggerMiddleware struct {
	thresholdDurationForLoggingRequest time.Duration
}

// NewRequestLoggerMiddleware returns a new instance of requestLoggerMiddleware
func NewRequestLoggerMiddleware(thresholdDurationForLoggingRequest time.Duration) *requestLoggerMiddleware {
	return &requestLoggerMiddleware{
		thresholdDurationForLoggingRequest: thresholdDurationForLoggingRequest,
	}
}

// MiddlewareHandlerFunc logs details about a request if it is not successful, or its duration is higher than a threshold
func (rlm *requestLoggerMiddleware) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Next()

		duration := time.Since(t)
		status := c.Writer.Status()

		shouldLogRequest := duration > rlm.thresholdDurationForLoggingRequest || status != http.StatusOK
		if !shouldLogRequest {
			return
		}

		log.Debug(computeLogTitle(status),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"duration", duration,
			"status", status,
		)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rlm *requestLoggerMiddleware) IsInterfaceNil() bool {
	return rlm == nil
}

func computeLogTitle(status int) string {
	logPrefix := prefixDurationTooLong
	if status == http.StatusBadRequest {
		logPrefix = prefixBadRequest
	} else if status == http.StatusInternalServerError {
		logPrefix = prefixInternalError
	} else if status != http.StatusOK {
		logPrefix = fmt.Sprintf("http code %d", status)
	}

	return fmt.Sprintf("%s api request", logPrefix)
}
