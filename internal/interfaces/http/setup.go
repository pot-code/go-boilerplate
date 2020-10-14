package http

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/domain"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"github.com/pot-code/go-boilerplate/internal/interfaces/http/middleware"
	"go.elastic.co/apm/module/apmechov4"
	"go.uber.org/zap"
)

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
	env := option.Env

	rdb := driver.NewRedisClient(option.KVStore.Host, option.KVStore.Port, option.KVStore.Password)
	jwtUtil := auth.NewJWTUtil(option.Security.JWTMethod, option.Security.JWTSecret, option.Security.TokenName, option.SessionTimeout)
	validator := infra.NewValidator()
	JWTMiddleware := middleware.VerifyToken(jwtUtil, &middleware.ValidateTokenOption{
		InBlackList: func(token string) (bool, error) {
			return rdb.Exists(token)
		},
	})
	RefreshMiddleware := middleware.RefreshToken(jwtUtil)

	app.HTTPErrorHandler = func(err error, c echo.Context) {
		if v, ok := err.(*echo.HTTPError); ok {
			c.String(v.Code, fmt.Sprintf("%v", v.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, infra.RESTStandardError{
			Code:  http.StatusInternalServerError,
			Title: err.Error(),
		})
	}
	app.Use(middleware.PanicHandling(
		&middleware.PanicHandlingOption{
			Logger: logger,
		},
	))
	app.Use(middleware.Logging(logger))
	if option.DevOP.APM {
		app.Use(apmechov4.Middleware())
	}
	if env == "development" {
		app.Use(middleware.CORSMiddleware)
	}
	app.Use(middleware.AbortRequest(&middleware.AbortRequestOption{
		Timeout: 30 * time.Second,
	}))
	app.Use(middleware.NoRouteMatched())

	UserHandler := NewUserHandler(jwtUtil,
		UserRepo, rdb, UserUserCase,
		option.Security.MaxLoginAttempts,
		option.Security.RetryTimeout,
		validator)
	LessonHandler := NewLessonHandler(LessonUseCase, jwtUtil)
	TimeSpentHandler := NewTimeSpentHandler(TimeSpentUseCase, jwtUtil, validator)

	expvarHandler := expvar.Handler()
	app.GET("/debug/vars", func(c echo.Context) error {
		expvarHandler.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}, JWTMiddleware)

	v1 := app.Group("/api/v1")
	TimeSpentGroup := v1.Group("/time-spent", JWTMiddleware, RefreshMiddleware)
	UserGroup := v1.Group("/user")
	LessonGroup := v1.Group("/lesson", JWTMiddleware, RefreshMiddleware)

	TimeSpentGroup.GET("/", TimeSpentHandler.HandleGetTimeSpent)

	UserGroup.POST("/login", UserHandler.HandleSignIn)
	UserGroup.GET("/sign-out", UserHandler.HandleSignOut)
	UserGroup.POST("/sign-up", UserHandler.HandleSignUp)
	UserGroup.GET("/exists", UserHandler.HandleUserExists)

	LessonGroup.GET("/progress", LessonHandler.HandleGetLessonProgress)

	v1.GET("/ws/echo", infra.WithHeartbeat(HandleEcho))

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