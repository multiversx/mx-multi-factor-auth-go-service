package testscommon

import (
	"github.com/gin-gonic/gin"
	"github.com/multiversx/multi-factor-auth-go-service/api/middleware"
)

type MiddlewareStub struct {
	UserAddress string
}

// MiddlewareHandlerFunc -
func (stub *MiddlewareStub) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.UserAddressKey, stub.UserAddress)
	}
}

// IsInterfaceNil -
func (stub *MiddlewareStub) IsInterfaceNil() bool {
	return stub == nil
}
