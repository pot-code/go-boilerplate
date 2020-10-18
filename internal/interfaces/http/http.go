package http

import (
	"expvar"
	"fmt"
	"log"
	"strings"

	"github.com/labstack/echo/v4"
	echo_middleware "github.com/labstack/echo/v4/middleware"
	"github.com/pot-code/go-boilerplate/internal/domain"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"github.com/pot-code/go-boilerplate/internal/interfaces/http/middleware"
	"go.elastic.co/apm/module/apmechov4"
	"go.uber.org/zap"
)

type endpoint struct {
	apiVersion string
	groups     []*apiGroup
}

type apiGroup struct {
	prefix      string
	middlewares []echo.MiddlewareFunc
	routes      []*route
}

type route struct {
	method      string
	path        string
	handler     echo.HandlerFunc
	middlewares []echo.MiddlewareFunc
}

// Serve create http transport server
func Serve(
	option *infra.AppConfig,
	UserUserCase domain.UserUseCase,
	UserRepo *auth.UserRepository,
	LessonUseCase domain.LessonUseCase,
	TimeSpentUseCase domain.TimeSpentUseCase,
	logger *zap.Logger,
) {
	app := echo.New()
	rdb := driver.NewRedisClient(option.KVStore.Host, option.KVStore.Port, option.KVStore.Password)
	jwtUtil := auth.NewJWTUtil(option.Security.JWTMethod,
		option.Security.JWTSecret,
		option.Security.TokenName,
		option.SessionTimeout)
	validator := infra.NewValidator()
	jwtMiddleware := middleware.VerifyToken(jwtUtil, &middleware.ValidateTokenOption{
		InBlackList: func(token string) (bool, error) {
			return rdb.Exists(token)
		},
	})
	refreshMiddleware := middleware.RefreshToken(jwtUtil)

	app.Use(echo_middleware.RequestID())
	app.Use(middleware.Logging(logger))
	app.Use(middleware.SetTraceLogger(logger))
	app.Use(middleware.ErrorHandling(
		&middleware.ErrorHandlingOption{
			LoggerKey: infra.ContextLoggerKey,
		},
	))
	app.Use(echo_middleware.Secure())
	if option.DevOP.APM {
		app.Use(apmechov4.Middleware())
	}
	app.Use(echo_middleware.CORS())
	app.Use(middleware.AbortRequest(&middleware.AbortRequestOption{
		Timeout: option.RequestTimeout,
	}))

	UserHandler := NewUserHandler(
		jwtUtil, UserRepo, rdb, UserUserCase,
		option.Security.MaxLoginAttempts,
		option.Security.RetryTimeout,
		validator,
	)
	LessonHandler := NewLessonHandler(LessonUseCase, jwtUtil)
	TimeSpentHandler := NewTimeSpentHandler(TimeSpentUseCase, jwtUtil, validator)

	expvarHandler := expvar.Handler()
	app.GET("/debug/vars", func(c echo.Context) error {
		expvarHandler.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}, jwtMiddleware)

	v1Endpoint := &endpoint{
		apiVersion: "api/v1",
		groups: []*apiGroup{
			{
				prefix: "/user",
				routes: []*route{
					{"POST", "/login", UserHandler.HandleSignIn, nil},
					{"PUT", "/sign-out", UserHandler.HandleSignOut, nil},
					{"POST", "/sign-up", UserHandler.HandleSignUp, nil},
					{"GET", "/exists", UserHandler.HandleUserExists, nil},
				},
			},
			{
				prefix:      "/lesson",
				middlewares: []echo.MiddlewareFunc{jwtMiddleware, refreshMiddleware},
				routes: []*route{
					{"GET", "/progress", LessonHandler.HandleGetLessonProgress, nil},
				},
			},
			{
				prefix:      "/time-spent",
				middlewares: []echo.MiddlewareFunc{jwtMiddleware, refreshMiddleware},
				routes: []*route{
					{"GET", "/", TimeSpentHandler.HandleGetTimeSpent, nil},
				},
			},
			{
				prefix: "/ws",
				routes: []*route{
					{"GET", "/echo", infra.WithHeartbeat(HandleEcho), nil},
				},
			},
		},
	}

	createEndpoint(app, v1Endpoint)

	printRoutes(app, logger)

	if err := app.Start(fmt.Sprintf("%s:%d", option.Host, option.Port)); err != nil {
		log.Fatal(err)
	}
}

func printRoutes(app *echo.Echo, logger *zap.Logger) {
	for _, route := range app.Routes() {
		if !strings.HasPrefix(route.Name, "github.com/labstack/echo") {
			name := route.Name
			trimIndex := strings.LastIndexByte(name, '/')
			logger.Debug("Registered route", zap.String("method", route.Method), zap.String("path", route.Path), zap.String("name", string(name[trimIndex+1:])))
		}
	}
}

func createEndpoint(app *echo.Echo, def *endpoint) {
	type RESTMethod func(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route

	var root *echo.Group
	if strings.HasPrefix(def.apiVersion, "/") {
		root = app.Group(def.apiVersion)
	} else {
		root = app.Group("/" + def.apiVersion)
	}

	for _, group := range def.groups {
		echoGroup := root.Group(group.prefix, group.middlewares...)
		for _, api := range group.routes {
			var method RESTMethod
			switch api.method {
			case "GET":
				method = echoGroup.GET
			case "POST":
				method = echoGroup.POST
			case "PUT":
				method = echoGroup.PUT
			case "DELETE":
				method = echoGroup.DELETE
			case "HEAD":
				method = echoGroup.HEAD
			case "CONNECT":
				method = echoGroup.CONNECT
			default:
				panic(fmt.Errorf("createEndpoint: unknown method %s", api.method))
			}
			method(api.path, api.handler, api.middlewares...)
		}
	}
}
