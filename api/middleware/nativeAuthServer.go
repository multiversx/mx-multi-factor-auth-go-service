package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/authentication"
	"github.com/gin-gonic/gin"
)

type nativeAuthServer struct {
	validator authentication.AuthServer
}

func NewNativeAuthServer(validator authentication.AuthServer) *nativeAuthServer {
	return &nativeAuthServer{
		validator: validator,
	}
}

func (middleware *nativeAuthServer) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := strings.Split(c.Request.Header.Get("Authorization"), "Bearer ")

		if len(authHeader) != 2 {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: errors.New("malformed token").Error(),
					Code:  shared.ReturnCodeRequestError,
				},
			)
			return
		}

		jwtToken := authHeader[1]
		err := middleware.validator.Validate(jwtToken)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: err.Error(),
					Code:  shared.ReturnCodeRequestError,
				},
			)
			return
		}

		c.Next()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (middleware *nativeAuthServer) IsInterfaceNil() bool {
	return middleware == nil
}
