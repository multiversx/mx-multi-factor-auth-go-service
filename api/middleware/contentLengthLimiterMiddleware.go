package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/api/shared"
)

type contentLengthLimiterMiddleware struct {
	MaxSizeBytes int64
}

// NewContentLengthLimiterMiddleware will abort all requests that have Content-Length size bigger
// than the one specified in config.
func NewContentLengthLimiterMiddleware(maxSizeBytes uint64) (*contentLengthLimiterMiddleware, error) {
	if maxSizeBytes < 10 {
		return nil, ErrMaxSizeByteTooSmall
	}
	return &contentLengthLimiterMiddleware{MaxSizeBytes: int64(maxSizeBytes)}, nil
}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests.
func (r *contentLengthLimiterMiddleware) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		size := c.Request.ContentLength

		if size == -1 {
			log.Debug("unknown content length", "error", errors.New("content length is -1"))
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: fmt.Errorf("%w, cannot process request", ErrUnknownContentLength).Error(),
					Code:  shared.ReturnCodeRequestError,
				},
			)
			return
		}

		if size > r.MaxSizeBytes {
			log.Debug("content length too large", "error", ErrContentLengthTooLarge)
			c.AbortWithStatusJSON(
				http.StatusRequestEntityTooLarge,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: fmt.Errorf("%w, cannot process request", ErrContentLengthTooLarge).Error(),
					Code:  shared.ReturnCodeRequestError,
				},
			)
			return
		}

		c.Next()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *contentLengthLimiterMiddleware) IsInterfaceNil() bool {
	return r == nil
}
