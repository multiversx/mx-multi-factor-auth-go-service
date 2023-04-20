package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apiErrors "github.com/multiversx/multi-factor-auth-go-service/api/errors"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/shared"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/authentication"
)

// UserAddressKey is the key of pair for the user address stored in the context map
const UserAddressKey = "userAddress"

var log = logger.GetOrCreate("middleware")

type ArgNativeAuth struct {
	Validator        authentication.AuthServer
	TokenHandler     authentication.AuthTokenHandler
	WhitelistHandler core.NativeAuthWhitelistHandler
}
type nativeAuth struct {
	validator        authentication.AuthServer
	tokenHandler     authentication.AuthTokenHandler
	whitelistHandler core.NativeAuthWhitelistHandler
}

// NewNativeAuth returns a new instance of nativeAuth
func NewNativeAuth(args ArgNativeAuth) (*nativeAuth, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &nativeAuth{
		validator:        args.Validator,
		tokenHandler:     args.TokenHandler,
		whitelistHandler: args.WhitelistHandler,
	}, nil
}

func checkArgs(args ArgNativeAuth) error {
	if check.IfNil(args.Validator) {
		return apiErrors.ErrNilNativeAuthServer
	}
	if check.IfNil(args.TokenHandler) {
		return authentication.ErrNilTokenHandler
	}
	if check.IfNil(args.WhitelistHandler) {
		return apiErrors.ErrNilNativeAuthWhitelistHandler
	}

	return nil
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
	isPostRequest := c.Request.Method == http.MethodPost
	isWhitelisted := middleware.whitelistHandler.IsWhitelisted(c.Request.URL.Path)
	return isPostRequest && !isWhitelisted
}

// IsInterfaceNil returns true if there is no value under the interface
func (middleware *nativeAuth) IsInterfaceNil() bool {
	return middleware == nil
}
