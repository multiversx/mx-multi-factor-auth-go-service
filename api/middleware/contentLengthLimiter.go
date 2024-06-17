package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/api/shared"

	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
)

const (
	unknownContentLengthSize = -1
)

type contentLengthLimiter struct {
	maxContentLengths map[string]uint64
}

// NewContentLengthLimiter will abort all requests that have Content-Length size bigger
// than the one specified in config.
func NewContentLengthLimiter(apiPackages map[string]config.APIPackageConfig) (*contentLengthLimiter, error) {
	maxContentLengths := map[string]uint64{}
	for group, groupCfg := range apiPackages {
		groupPath := fmt.Sprintf("/%s", group)

		for _, r := range groupCfg.Routes {
			fullPath := fmt.Sprintf("%s%s", groupPath, r.Name)
			maxContentLengths[fullPath] = r.MaxContentLength
		}
	}

	return &contentLengthLimiter{
		maxContentLengths: maxContentLengths,
	}, nil

}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests.
func (r *contentLengthLimiter) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// no need to check if the request is GET as the Content-Length will be 0.
		if c.Request.Method == http.MethodGet {
			return
		}

		size := c.Request.ContentLength
		maxSizeBytes, ok := r.maxContentLengths[c.Request.URL.Path]
		if !ok {
			log.Debug(fmt.Sprintf("invalid path: %s", c.Request.URL.Path))
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: fmt.Errorf("%w, cannot process request", ErrInvalidPath).Error(),
					Code:  shared.ReturnCodeRequestError,
				},
			)
			return
		}

		if size == unknownContentLengthSize {
			log.Debug(fmt.Sprintf("received -1 content length: %s", ErrUnknownContentLength.Error()))
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

		if size > int64(maxSizeBytes) {
			log.Debug(fmt.Sprintf("%s, received %d, max allowed %d", ErrContentLengthTooLarge.Error(), size, maxSizeBytes))
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
func (r *contentLengthLimiter) IsInterfaceNil() bool {
	return r == nil
}
