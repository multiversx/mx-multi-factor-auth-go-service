package gin

import (
	"errors"
	"net/http"
	"testing"
	"time"

	apiErrors "github.com/multiversx/multi-factor-auth-go-service/api/errors"
	"github.com/multiversx/multi-factor-auth-go-service/api/shared"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon/facade"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon/groups"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/middleware"
	"github.com/multiversx/mx-sdk-go/authentication/native/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArgsNewWebServer() ArgsNewWebServer {
	return ArgsNewWebServer{
		Config: config.Configs{
			FlagsConfig: config.ContextFlagsConfig{
				RestApiInterface: "127.0.0.1:8080",
				EnablePprof:      true,
			},
			ApiRoutesConfig: config.ApiRoutesConfig{
				Logging: config.ApiLoggingConfig{
					LoggingEnabled:          true,
					ThresholdInMicroSeconds: 10,
				},
				APIPackages: make(map[string]config.APIPackageConfig),
			},
			GeneralConfig: config.Config{
				Antiflood: config.AntifloodConfig{
					Enabled: true,
					WebServer: config.WebServerAntifloodConfig{
						SimultaneousRequests:         1,
						SameSourceRequests:           1,
						SameSourceResetIntervalInSec: 1,
					},
				},
			},
		},
		Facade:       &facade.GuardianFacadeStub{},
		AuthServer:   &mock.AuthServerStub{},
		TokenHandler: &mock.AuthTokenHandlerStub{},
	}
}

func TestNewWebServerHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsNewWebServer()
		args.Facade = nil

		ws, err := NewWebServerHandler(args)
		assert.Equal(t, apiErrors.ErrNilFacade, err)
		assert.True(t, check.IfNil(ws))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ws, err := NewWebServerHandler(createMockArgsNewWebServer())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ws))
	})
}

func TestWebServer_StartHttpServer(t *testing.T) {
	t.Run("RestApiInterface returns WebServerOffString", func(t *testing.T) {
		args := createMockArgsNewWebServer()
		args.Config.ApiRoutesConfig.RestApiInterface = core.WebServerOffString

		ws, _ := NewWebServerHandler(args)
		assert.False(t, check.IfNil(ws))

		err := ws.StartHttpServer()
		assert.Nil(t, err)
	})
	t.Run("createMiddlewareLimiters returns error due to middleware.NewSourceThrottler error", func(t *testing.T) {
		args := createMockArgsNewWebServer()
		args.Config.GeneralConfig.Antiflood = config.AntifloodConfig{
			Enabled: true,
			WebServer: config.WebServerAntifloodConfig{
				SimultaneousRequests:         1,
				SameSourceRequests:           0,
				SameSourceResetIntervalInSec: 1,
			},
		}
		ws, _ := NewWebServerHandler(args)
		assert.False(t, check.IfNil(ws))

		err := ws.StartHttpServer()
		assert.Equal(t, middleware.ErrInvalidMaxNumRequests, err)
	})
	t.Run("createMiddlewareLimiters returns error due to middleware.NewGlobalThrottler error", func(t *testing.T) {
		args := createMockArgsNewWebServer()
		args.Config.GeneralConfig.Antiflood = config.AntifloodConfig{
			Enabled: true,
			WebServer: config.WebServerAntifloodConfig{
				SimultaneousRequests:         0,
				SameSourceRequests:           1,
				SameSourceResetIntervalInSec: 1,
			},
		}
		ws, _ := NewWebServerHandler(args)
		assert.False(t, check.IfNil(ws))

		err := ws.StartHttpServer()
		assert.Equal(t, middleware.ErrInvalidMaxNumRequests, err)
	})
	t.Run("upgrade on get returns error", func(t *testing.T) {
		ws, _ := NewWebServerHandler(createMockArgsNewWebServer())
		assert.False(t, check.IfNil(ws))

		err := ws.StartHttpServer()
		assert.Nil(t, err)

		time.Sleep(2 * time.Second)

		resp, err := http.Get("http://127.0.0.1:8080/log")
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // Bad request

		time.Sleep(2 * time.Second)
		err = ws.Close()
		assert.Nil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		ws, _ := NewWebServerHandler(createMockArgsNewWebServer())
		assert.False(t, check.IfNil(ws))

		err := ws.StartHttpServer()
		assert.Nil(t, err)

		time.Sleep(2 * time.Second)

		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080/log", nil)
		assert.Nil(t, err)

		req.Header.Set("Sec-Websocket-Version", "13")
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-Websocket-Key", "key")

		resp, err := client.Do(req)
		assert.Nil(t, err)

		err = resp.Body.Close()
		assert.Nil(t, err)

		time.Sleep(2 * time.Second)
		err = ws.Close()
		assert.Nil(t, err)
	})
}

func TestWebServer_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("update with nil facade", func(t *testing.T) {
		t.Parallel()

		ws, _ := NewWebServerHandler(createMockArgsNewWebServer())
		assert.False(t, check.IfNil(ws))

		err := ws.UpdateFacade(nil)
		assert.Equal(t, apiErrors.ErrNilFacade, err)
	})
	t.Run("should work - one of the groupHandlers returns err", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsNewWebServer()
		args.Config.ApiRoutesConfig.RestApiInterface = "provided interface"

		ws, _ := NewWebServerHandler(args)
		assert.False(t, check.IfNil(ws))

		ws.groups = make(map[string]shared.GroupHandler)
		ws.groups["first"] = &groups.GroupHandlerStub{
			UpdateFacadeCalled: func(newFacade shared.FacadeHandler) error {
				return errors.New("error")
			},
		}
		ws.groups["second"] = &groups.GroupHandlerStub{
			UpdateFacadeCalled: func(newFacade shared.FacadeHandler) error {
				return nil
			},
		}

		err := ws.UpdateFacade(&facade.GuardianFacadeStub{})
		assert.Nil(t, err)
	})
}
