package main

import (
	"log"

	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/uuid"
	ihttp "github.com/pot-code/go-boilerplate/internal/interfaces/http"
	"github.com/pot-code/go-boilerplate/internal/lesson"
	"github.com/pot-code/go-boilerplate/internal/time_spent"
	"github.com/pot-code/go-boilerplate/internal/user"
	"go.uber.org/zap"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	option, err := infra.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := infra.NewLogger(&infra.LoggingConfig{
		FilePath: option.Logging.FilePath,
		Level:    option.Logging.Level,
		AppID:    option.AppID,
		Env:      option.Env,
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %s\n", err)
	}
	logger = logger.With(
		zap.String("service.id", option.AppID),
	)
	defer logger.Sync()

	dbConn, err := driver.GetDBConnection(&driver.DBConfig{
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
	logger.Debug("Create mysql connection instance", zap.String("db.driver", option.Database.Driver),
		zap.String("db.schema", option.Database.Schema),
		zap.String("db.host", option.Database.Host),
		zap.Any("config", option.Database),
	)

	UUIDGenerator := uuid.NewNanoIDGenerator(option.Security.IDLength)
	UserRepo := user.NewUserRepository(dbConn, UUIDGenerator)
	UserUserCase := user.NewUserUseCase(UserRepo)

	LessonRepo := lesson.NewLessonRepository(dbConn)
	LessonUseCase := lesson.NewLessonUseCase(LessonRepo)

	TimeSpentRepo := time_spent.NewTimeSpentRepository(dbConn)
	TimeSpentUseCase := time_spent.NewTimeSpentUseCase(TimeSpentRepo)

	ihttp.Serve(dbConn, option, UserUserCase, UserRepo, LessonUseCase, TimeSpentUseCase, logger)
}
