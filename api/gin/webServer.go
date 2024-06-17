package gin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/api/logs"
	"github.com/multiversx/mx-chain-go/api/middleware"
	chainShared "github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/facade"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/authentication"

	apiErrors "github.com/multiversx/mx-multi-factor-auth-go-service/api/errors"
	"github.com/multiversx/mx-multi-factor-auth-go-service/api/groups"
	mfaMiddleware "github.com/multiversx/mx-multi-factor-auth-go-service/api/middleware"
	"github.com/multiversx/mx-multi-factor-auth-go-service/api/shared"
	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
)

var log = logger.GetOrCreate("api")

// ArgsNewWebServer holds the arguments needed to create a new instance of webServer
type ArgsNewWebServer struct {
	Facade                     shared.FacadeHandler
	Config                     config.Configs
	AuthServer                 authentication.AuthServer
	TokenHandler               authentication.AuthTokenHandler
	NativeAuthWhitelistHandler core.NativeAuthWhitelistHandler
	StatusMetricsHandler       core.StatusMetricsHandler
}

type webServer struct {
	sync.RWMutex
	facade                     shared.FacadeHandler
	config                     config.Configs
	authServer                 authentication.AuthServer
	tokenHandler               authentication.AuthTokenHandler
	nativeAuthWhitelistHandler core.NativeAuthWhitelistHandler
	httpServer                 chainShared.HttpServerCloser
	statusMetrics              core.StatusMetricsHandler
	groups                     map[string]shared.GroupHandler
	cancelFunc                 func()
}

// NewWebServerHandler returns a new instance of webServer
func NewWebServerHandler(args ArgsNewWebServer) (*webServer, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	gws := &webServer{
		facade:                     args.Facade,
		config:                     args.Config,
		authServer:                 args.AuthServer,
		tokenHandler:               args.TokenHandler,
		nativeAuthWhitelistHandler: args.NativeAuthWhitelistHandler,
		statusMetrics:              args.StatusMetricsHandler,
	}

	return gws, nil
}

// checkArgs check the arguments of an ArgsNewWebServer
func checkArgs(args ArgsNewWebServer) error {

	if check.IfNil(args.Facade) {
		return apiErrors.ErrNilFacade
	}
	if check.IfNil(args.AuthServer) {
		return apiErrors.ErrNilNativeAuthServer
	}
	if check.IfNil(args.TokenHandler) {
		return authentication.ErrNilTokenHandler
	}
	if check.IfNil(args.NativeAuthWhitelistHandler) {
		return apiErrors.ErrNilNativeAuthWhitelistHandler
	}
	if check.IfNil(args.StatusMetricsHandler) {
		return core.ErrNilMetricsHandler
	}

	return nil
}

// StartHttpServer will create a new instance of http.Server and populate it with all the routes
func (ws *webServer) StartHttpServer() error {
	ws.Lock()
	defer ws.Unlock()

	apiInterface := ws.config.ApiRoutesConfig.RestApiInterface
	if ws.config.FlagsConfig.RestApiInterface != facade.DefaultRestInterface {
		apiInterface = ws.config.FlagsConfig.RestApiInterface
	}

	if apiInterface == core.WebServerOffString {
		log.Debug("web server is turned off")
		return nil
	}

	var engine *gin.Engine

	gin.DefaultWriter = &ginWriter{}
	gin.DefaultErrorWriter = &ginErrorWriter{}
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	engine = gin.Default()
	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AddAllowHeaders("Authorization")
	engine.Use(cors.New(cfg))

	err := ws.setOptionsForClientIP(engine)
	if err != nil {
		return err
	}

	if ws.config.FlagsConfig.StartSwaggerUI {
		engine.Use(static.ServeRoot("/", "swagger/ui"))
	}

	err = ws.createGroups()
	if err != nil {
		return err
	}

	processors, err := ws.createMiddlewareLimiters()
	if err != nil {
		return err
	}

	for idx, proc := range processors {
		if check.IfNil(proc) {
			log.Error("got nil middleware processor, skipping it...", "index", idx)
			continue
		}

		engine.Use(proc.MiddlewareHandlerFunc())
	}

	ws.registerRoutes(engine)

	server := &http.Server{Addr: apiInterface, Handler: engine}
	log.Debug("creating gin web sever", "interface", apiInterface)
	ws.httpServer, err = NewHttpServer(server)
	if err != nil {
		return err
	}

	log.Debug("starting web server")
	go ws.httpServer.Start()

	return nil
}

func (ws *webServer) createGroups() error {
	groupsMap := make(map[string]shared.GroupHandler)

	guardianGroup, err := groups.NewGuardianGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["guardian"] = guardianGroup

	statusGroup, err := groups.NewStatusGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["status"] = statusGroup

	ws.groups = groupsMap

	return nil
}

// UpdateFacade will update webServer facade.
func (ws *webServer) UpdateFacade(facade shared.FacadeHandler) error {
	if check.IfNil(facade) {
		return apiErrors.ErrNilFacade
	}

	ws.Lock()
	defer ws.Unlock()

	ws.facade = facade

	for groupName, groupHandler := range ws.groups {
		log.Debug("upgrading facade for gin API group", "group name", groupName)
		err := groupHandler.UpdateFacade(facade)
		if err != nil {
			log.Error("cannot update facade for gin API group", "group name", groupName, "error", err)
		}
	}

	return nil
}

func (ws *webServer) setOptionsForClientIP(engine *gin.Engine) error {
	engine.ForwardedByClientIP = ws.config.ExternalConfig.Gin.ForwardedByClientIP

	engine.TrustedPlatform = ws.config.ExternalConfig.Gin.TrustedPlatform

	remoteIPHeaders := ws.config.ExternalConfig.Gin.RemoteIPHeaders
	if len(remoteIPHeaders) != 0 {
		engine.RemoteIPHeaders = remoteIPHeaders
	}

	trustedProxies := ws.config.ExternalConfig.Gin.TrustedProxies
	if len(trustedProxies) == 0 {
		// disable trusted proxies checking
		// will get IP directly from `RemoteAddr`, since headers are not trustworthy
		return engine.SetTrustedProxies(nil)
	}

	return engine.SetTrustedProxies(trustedProxies)
}

func (ws *webServer) registerRoutes(ginRouter *gin.Engine) {

	for groupName, groupHandler := range ws.groups {
		log.Debug("registering gin API group", "group name", groupName)
		ginGroup := ginRouter.Group(fmt.Sprintf("/%s", groupName))
		groupHandler.RegisterRoutes(ginGroup, ws.config.ApiRoutesConfig)
	}

	marshallerForLogs := &marshal.GogoProtoMarshalizer{}
	registerLoggerWsRoute(ginRouter, marshallerForLogs)

	if ws.config.FlagsConfig.EnablePprof {
		pprof.Register(ginRouter)
	}
}

// registerLoggerWsRoute will register the log route
func registerLoggerWsRoute(ws *gin.Engine, marshaller marshal.Marshalizer) {
	upgrader := websocket.Upgrader{}

	ws.GET("/log", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls, err := logs.NewLogSender(marshaller, conn, log)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls.StartSendingBlocking()
	})
}

func (ws *webServer) createMiddlewareLimiters() ([]chainShared.MiddlewareProcessor, error) {
	middlewares := make([]chainShared.MiddlewareProcessor, 0)

	metricsMiddleware, err := mfaMiddleware.NewMetricsMiddleware(ws.statusMetrics, ws.config.ApiRoutesConfig)
	if err != nil {
		return nil, err
	}
	middlewares = append(middlewares, metricsMiddleware)

	if ws.config.ApiRoutesConfig.Logging.LoggingEnabled {
		responseLoggerMiddleware := middleware.NewResponseLoggerMiddleware(time.Duration(ws.config.ApiRoutesConfig.Logging.ThresholdInMicroSeconds) * time.Microsecond)
		middlewares = append(middlewares, responseLoggerMiddleware)
	}

	antifloodCfg := ws.config.GeneralConfig.Antiflood
	if antifloodCfg.Enabled {
		sourceLimiter, err := middleware.NewSourceThrottler(antifloodCfg.WebServer.SameSourceRequests)
		if err != nil {
			return nil, err
		}

		var ctx context.Context
		ctx, ws.cancelFunc = context.WithCancel(context.Background())

		go ws.sourceLimiterReset(ctx, sourceLimiter)

		middlewares = append(middlewares, sourceLimiter)

		globalLimiter, err := middleware.NewGlobalThrottler(antifloodCfg.WebServer.SimultaneousRequests)
		if err != nil {
			return nil, err
		}

		middlewares = append(middlewares, globalLimiter)
	}

	argsNativeAuth := mfaMiddleware.ArgNativeAuth{
		Validator:        ws.authServer,
		TokenHandler:     ws.tokenHandler,
		WhitelistHandler: ws.nativeAuthWhitelistHandler,
	}
	nativeAuthLimiter, err := mfaMiddleware.NewNativeAuth(argsNativeAuth)
	if err != nil {
		return nil, err
	}

	middlewares = append(middlewares, nativeAuthLimiter)

	userContextMiddleware := mfaMiddleware.NewUserContext()
	middlewares = append(middlewares, userContextMiddleware)

	m, err := mfaMiddleware.NewContentLengthLimiter(ws.config.ApiRoutesConfig.APIPackages)
	if err != nil {
		return nil, err
	}
	middlewares = append(middlewares, m)

	return middlewares, nil
}

func (ws *webServer) sourceLimiterReset(ctx context.Context, reset resetHandler) {
	betweenResetDuration := time.Second * time.Duration(ws.config.GeneralConfig.Antiflood.WebServer.SameSourceResetIntervalInSec)
	timer := time.NewTimer(betweenResetDuration)
	defer timer.Stop()

	for {
		timer.Reset(betweenResetDuration)

		select {
		case <-timer.C:
			log.Trace("calling reset on WS source limiter")
			reset.Reset()
		case <-ctx.Done():
			log.Debug("closing nodeFacade.sourceLimiterReset go routine")
			return
		}
	}
}

// Close will handle the closing of inner components
func (ws *webServer) Close() error {
	if ws.cancelFunc != nil {
		ws.cancelFunc()
	}

	var err error
	ws.Lock()
	if ws.httpServer != nil {
		err = ws.httpServer.Close()
	}
	ws.Unlock()

	if err != nil {
		err = fmt.Errorf("%w while closing the http server in gin/webServer", err)
	}

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ws *webServer) IsInterfaceNil() bool {
	return ws == nil
}
