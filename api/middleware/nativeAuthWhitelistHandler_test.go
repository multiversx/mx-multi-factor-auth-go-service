package middleware

import (
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/stretchr/testify/require"
)

func TestNativeAuthWhitelistHandler(t *testing.T) {
	t.Parallel()

	providedMap := map[string]config.APIPackageConfig{
		"guardian": {
			Routes: []config.RouteConfig{
				{
					Name: "/register",
					Open: true,
					Auth: true,
				},
				{
					Name: "/sign-transaction",
					Open: true,
					Auth: false,
				},
			},
		},
		"status": {
			Routes: []config.RouteConfig{
				{
					Name: "/check-status",
					Open: true,
					Auth: false,
				},
			},
		},
	}
	handler := NewNativeAuthWhitelistHandler(providedMap)
	require.NotNil(t, handler)

	require.True(t, handler.IsWhitelisted("/guardian/sign-transaction"))
	require.True(t, handler.IsWhitelisted("/status/check-status"))
	require.True(t, handler.IsWhitelisted("/guardian"))
	require.True(t, handler.IsWhitelisted("/status"))
	require.True(t, handler.IsWhitelisted("/log"))
	require.False(t, handler.IsWhitelisted("/guardian/register"))
	require.False(t, handler.IsWhitelisted("guardian/sign-transaction"))
	require.False(t, handler.IsWhitelisted("/sign-transaction"))
	require.False(t, handler.IsWhitelisted("guardian"))
	require.False(t, handler.IsWhitelisted(""))
}

func TestNativeAuthWhitelistHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var handler *nativeAuthWhitelistHandler
	require.True(t, handler.IsInterfaceNil())

	handler = NewNativeAuthWhitelistHandler(map[string]config.APIPackageConfig{})
	require.False(t, handler.IsInterfaceNil())
}
