package middleware

import (
	"net/http"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/authentication"
	apiErrors "github.com/ElrondNetwork/multi-factor-auth-go-service/api/errors"
	"github.com/gin-gonic/gin"
)

const UserAddressKey = "userAddress"

type nativeAuth struct {
	validator authentication.AuthServer
}

// NewNativeAuth returns a new instance of nativeAuth
func NewNativeAuth(validator authentication.AuthServer) (*nativeAuth, error) {
	if check.IfNil(validator) {
		return nil, apiErrors.ErrNilNativeAuthServer
	}
	return &nativeAuth{
		validator: validator,
	}, nil
}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests
func (middleware *nativeAuth) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware.checkIfGuarded(c)
		authHeader := strings.Split(c.Request.Header.Get("Authorization"), "Bearer ")

		if len(authHeader) != 2 {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: ErrMalformedToken.Error(),
					Code:  shared.ReturnCodeRequestError,
				},
			)
			return
		}

		jwtToken := authHeader[1]
		address, err := middleware.validator.Validate(jwtToken)
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

		c.Set(UserAddressKey, address)
		c.Next()
	}
}

func (middleware *nativeAuth) checkIfGuarded(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.Next()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (middleware *nativeAuth) IsInterfaceNil() bool {
	return middleware == nil
}
