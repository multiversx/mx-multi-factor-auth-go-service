package groups_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/multiversx/mx-multi-factor-auth-go-service/api/shared"
	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon"
)

type generalResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func startWebServer(group shared.GroupHandler, path string, apiConfig config.ApiRoutesConfig, userAddress string) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	if len(userAddress) > 0 {
		middlewareStub := testscommon.MiddlewareStub{
			UserAddress: userAddress,
		}
		ws.Use(middlewareStub.MiddlewareHandlerFunc())
	}
	routes := ws.Group(path)
	group.RegisterRoutes(routes, apiConfig)
	return ws
}

func getServiceRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"guardian": {
				Routes: []config.RouteConfig{
					{Name: "/register", Open: true},
					{Name: "/sign-message", Open: true},
					{Name: "/sign-transaction", Open: true},
					{Name: "/sign-multiple-transactions", Open: true},
					{Name: "/set-security-mode", Open: true},
					{Name: "/unset-security-mode", Open: true},
					{Name: "/debug", Open: true},
					{Name: "/verify-code", Open: true},
					{Name: "/registered-users", Open: true},
					{Name: "/config", Open: true},
				},
			},
		},
	}
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	logError(err)
}

func requestToReader(request interface{}) io.Reader {
	data, _ := json.Marshal(request)
	return bytes.NewReader(data)
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
