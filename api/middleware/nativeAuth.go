package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apiErrors "github.com/multiversx/multi-factor-auth-go-service/api/errors"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/shared"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/authentication"
)

// UserAddressKey is the key of pair for the user address stored in the context map
const UserAddressKey = "userAddress"

var log = logger.GetOrCreate("middleware")

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
			log.Trace("cannot parse JWT token", "error", ErrMalformedToken.Error())
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
			log.Trace("cannot decode JWT Token", "error", err.Error(), "token", jwtToken)
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
			log.Trace("JWT Token validation failed", "reason", err.Error(), "token", jwtToken)
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
