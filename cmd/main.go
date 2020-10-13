package main

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	repo "github.com/pot-code/go-boilerplate/internal/infrastructure/repository"
	ihttp "github.com/pot-code/go-boilerplate/internal/interfaces/http"
	"github.com/pot-code/go-boilerplate/internal/interfaces/http/middleware"
	"github.com/pot-code/go-boilerplate/internal/usecase"
	"go.elastic.co/apm/module/apmechov4"
	"go.uber.org/zap"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	if cfg, err := InitConfig(); err == nil {
		httpServer(cfg)
	} else {
		log.Fatal(err)
	}
}

func httpServer(option *AppConfig) {
	env := option.Env
	app := echo.New()

	logger, err := infra.NewLogger(&infra.LoggingConfig{
		FilePath: option.Logging.FilePath,
		Level:    option.Logging.Level,
		AppID:    option.AppID,
		Env:      option.Env,
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %s\n", err)
	}
	defer logger.Sync()

	dbConn, err := driver.GetDBConnection(&driver.DBConfig{
		Logger:   logger,
		User:     option.Database.User,
		Password: option.Database.Password,
		MaxConn:  option.Database.MaxConn,
		Protocol: option.Database.Protocol,
		Driver:   option.Database.Driver,
		Host:     option.Database.Host,
		Port:     option.Database.Port,
		Query:    option.Database.Query,
		Schema:   option.Database.Schema,
	})
	if err != nil {
		log.Fatalf("Failed to create DB connection: %s\n", err)
	}

	rdb := driver.NewRedisClient(option.KVStore.Host, option.KVStore.Port, option.KVStore.Password)
	jwtUtil := auth.NewJWTUtil(option.Security.JWTMethod, option.Security.JWTSecret, option.Security.TokenName, option.SessionTimeout)
	UUIDGenerator := infra.NewNanoIDGenerator(option.Security.IDLength)
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

	UserRepo := auth.NewUserRepository(dbConn, UUIDGenerator)
	UserUserCase := auth.NewUserUseCase(UserRepo)
	UserHandler := ihttp.NewUserHandler(jwtUtil, UserRepo, rdb, UserUserCase, option.Security.MaxLoginAttempts, validator)

	LessonRepo := repo.NewLessonRepository(dbConn)
	LessonUseCase := usecase.NewLessonUseCase(LessonRepo)
	LessonHandler := ihttp.NewLessonHandler(LessonUseCase, jwtUtil)

	TimeSpentRepo := repo.NewTimeSpentRepository(dbConn)
	TimeSpentUseCase := usecase.NewTimeSpentUseCase(TimeSpentRepo)
	TimeSpentHandler := ihttp.NewTimeSpentHandler(TimeSpentUseCase, jwtUtil, validator)

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

	v1.GET("/ws/echo", infra.WithHeartbeat(ihttp.HandleEcho))

	if option.Env == "development" {
		printRoutes(app, logger)
	}
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
