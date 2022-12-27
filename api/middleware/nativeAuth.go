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

// UserAddressKey is the key of pair for the user address stored in the context map
const UserAddressKey = "userAddress"

type nativeAuth struct {
	validator    authentication.AuthServer
	tokenHandler authentication.AuthTokenHandler
}

// NewNativeAuth returns a new instance of nativeAuth
func NewNativeAuth(validator authentication.AuthServer, tokenHandler authentication.AuthTokenHandler) (*nativeAuth, error) {
	if check.IfNil(validator) {
		return nil, apiErrors.ErrNilNativeAuthServer
	}
	if check.IfNil(tokenHandler) {
		return nil, authentication.ErrNilTokenHandler
	}
	return &nativeAuth{
		validator:    validator,
		tokenHandler: tokenHandler,
	}, nil
}

// MiddlewareHandlerFunc returns the handler func used by the gin server when processing requests
func (middleware *nativeAuth) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.checkIfGuarded(c) {
			c.Next()
			return
		}
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
		authToken, err := middleware.tokenHandler.Decode(jwtToken)
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

		err = middleware.validator.Validate(authToken)
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

		c.Set(UserAddressKey, string(authToken.GetAddress()))
		c.Next()
	}
}

func (middleware *nativeAuth) checkIfGuarded(c *gin.Context) bool {
	return c.Request.Method == http.MethodPost
}

// IsInterfaceNil returns true if there is no value under the interface
func (middleware *nativeAuth) IsInterfaceNil() bool {
	return middleware == nil
}
