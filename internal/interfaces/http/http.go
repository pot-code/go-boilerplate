package http

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/labstack/echo/v4"
	echo_middleware "github.com/labstack/echo/v4/middleware"
	"github.com/pot-code/go-boilerplate/internal/domain"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/validate"
	"github.com/pot-code/go-boilerplate/internal/interfaces/http/middleware"
	"go.elastic.co/apm/module/apmechov4"
	"go.uber.org/zap"
)

type endpoint struct {
	apiVersion  string
	middlewares []echo.MiddlewareFunc
	groups      []*apiGroup
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
	conn driver.ITransactionalDB,
	rdb driver.KeyValueDB,
	option *infra.AppConfig,
	UserUserCase domain.UserUseCase,
	UserRepo domain.UserRepository,
	LessonUseCase domain.LessonUseCase,
	TimeSpentUseCase domain.TimeSpentUseCase,
	logger *zap.Logger,
) {
	app := echo.New()
	jwtUtil := auth.NewJWTUtil(option.Security.JWTMethod,
		option.Security.JWTSecret,
		option.Security.TokenName,
		option.SessionTimeout)
	validator := validate.NewValidator()
	websocket := infra.NewWebsocket()
	jwtMiddleware := middleware.VerifyToken(jwtUtil, &middleware.ValidateTokenOption{
		InBlackList: func(token string) (bool, error) {
			return rdb.Exists(token)
		},
	})
	refreshMiddleware := middleware.RefreshToken(jwtUtil)

	registerLivenessProbe(app, conn, rdb)
	if option.Env == infra.EnvDevelopment {
		registerProfileEndpoints(app)
	}
	app.Use(middleware.Logging(logger, &middleware.LoggingConfig{
		Skipper: func(e echo.Context) bool {
			if strings.HasPrefix(e.Request().RequestURI, "/healthz") {
				return true
			}
			return false
		},
	}))
	app.Use(middleware.ErrorHandling(
		&middleware.ErrorHandlingOption{
			Handler: func(c echo.Context, err error) {
				traceID := c.Response().Header().Get(echo.HeaderXRequestID)
				c.JSON(http.StatusInternalServerError,
					NewRESTStandardError(http.StatusInternalServerError, err.Error()).SetTraceID(traceID),
				)
				logger.Error(err.Error(), zap.String("trace.id", traceID))
			},
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
		jwtUtil, conn, UserRepo, rdb, UserUserCase,
		option.Security.MaxLoginAttempts,
		option.Security.RetryTimeout,
		validator,
	)
	LessonHandler := NewLessonHandler(LessonUseCase, jwtUtil)
	TimeSpentHandler := NewTimeSpentHandler(TimeSpentUseCase, jwtUtil, validator)

	createEndpoint(app, v1Endpoint(
		websocket,
		UserHandler,
		LessonHandler,
		TimeSpentHandler,
		jwtMiddleware, refreshMiddleware, echo_middleware.RequestID(), middleware.SetTraceLogger(logger),
	))

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

func registerLivenessProbe(app *echo.Echo, db driver.ITransactionalDB, rdb driver.KeyValueDB) {
	app.GET("/healthz", func(c echo.Context) error {
		if db.Ping() == nil && rdb.Ping() == nil {
			c.NoContent(http.StatusOK)
		} else {
			c.NoContent(http.StatusServiceUnavailable)
		}
		return nil
	})
}

func registerProfileEndpoints(app *echo.Echo) {
	expvarHandler := expvar.Handler()
	app.GET("/debug/vars", func(c echo.Context) error {
		expvarHandler.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	})
	app.GET("/debug/pprof/", func(c echo.Context) error {
		pprof.Index(c.Response().Writer, c.Request())
		return nil
	})
	app.GET("/debug/pprof/:name", func(c echo.Context) error {
		switch c.Param("name") {
		case "cmdline":
			pprof.Cmdline(c.Response().Writer, c.Request())
		case "profile":
			pprof.Profile(c.Response().Writer, c.Request())
		case "symbol":
			pprof.Symbol(c.Response().Writer, c.Request())
		case "trace":
			pprof.Trace(c.Response().Writer, c.Request())
		default:
			pprof.Handler(c.Param("name")).ServeHTTP(c.Response().Writer, c.Request())
		}
		return nil
	})
}

func createEndpoint(app *echo.Echo, def *endpoint) {
	type RESTMethod func(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route

	var root *echo.Group
	if strings.HasPrefix(def.apiVersion, "/") {
		root = app.Group(def.apiVersion, def.middlewares...)
	} else {
		root = app.Group("/"+def.apiVersion, def.middlewares...)
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
