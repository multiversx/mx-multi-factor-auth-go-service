package middleware

import (
	"fmt"
	"strings"

	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
)

type nativeAuthWhitelistHandler struct {
	whitelistedRoutesMap map[string]struct{}
}

// NewNativeAuthWhitelistHandler returns a new instance of nativeAuthWhitelistHandler
func NewNativeAuthWhitelistHandler(apiPackages map[string]config.APIPackageConfig) *nativeAuthWhitelistHandler {
	whitelistedRoutes := make(map[string]struct{})
	for group, groupCfg := range apiPackages {
		groupPath := fmt.Sprintf("/%s", group)
		whitelistedRoutes[groupPath] = struct{}{}

		for _, route := range groupCfg.Routes {
			if !route.Auth {
				fullPath := fmt.Sprintf("%s%s", groupPath, route.Name)
				basePath := trimPathPlaceholder(fullPath)
				whitelistedRoutes[basePath] = struct{}{}
			}
		}
	}
	whitelistedRoutes["/log"] = struct{}{}

	return &nativeAuthWhitelistHandler{
		whitelistedRoutesMap: whitelistedRoutes,
	}
}

func trimPathPlaceholder(path string) string {
	parts := strings.Split(path, ":")
	if len(parts) > 0 {
		return "/" + strings.Trim(parts[0], "/")
	}

	return path
}

func extractBaseRoutePath(path string) string {
	parts := strings.Split(path, "/")

	if len(parts) > 2 {
		return "/" + parts[1] + "/" + parts[2] // group and base path
	}

	return path
}

// IsWhitelisted returns true if the provided route is whitelisted for native authentication
func (handler *nativeAuthWhitelistHandler) IsWhitelisted(route string) bool {
	baseRoute := extractBaseRoutePath(route)
	_, found := handler.whitelistedRoutesMap[baseRoute]
	log.Error("adas", "map", handler.whitelistedRoutesMap, "bb", baseRoute)
	return found
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *nativeAuthWhitelistHandler) IsInterfaceNil() bool {
	return handler == nil
}
