package middleware

import (
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/config"
)

type nativeAuthWhitelistHandler struct {
	whitelistedRoutesMap map[string]struct{}
}

// NewNativeAuthWhitelistHandler returns a new instance of nativeAuthWhitelistHandler
func NewNativeAuthWhitelistHandler(apiPackages map[string]config.APIPackageConfig) *nativeAuthWhitelistHandler {
	whitelistedRoutes := make(map[string]struct{})
	for group, groupCfg := range apiPackages {
		for _, route := range groupCfg.Routes {
			if !route.Auth {
				fullPath := fmt.Sprintf("/%s%s", group, route.Name)
				whitelistedRoutes[fullPath] = struct{}{}
			}
		}
	}
	return &nativeAuthWhitelistHandler{
		whitelistedRoutesMap: whitelistedRoutes,
	}
}

// IsWhitelisted returns true if the provided route is whitelisted for native authentication
func (handler *nativeAuthWhitelistHandler) IsWhitelisted(route string) bool {
	_, found := handler.whitelistedRoutesMap[route]
	return found
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *nativeAuthWhitelistHandler) IsInterfaceNil() bool {
	return handler == nil
}
