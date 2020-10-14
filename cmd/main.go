package main

import (
	"log"

	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	repo "github.com/pot-code/go-boilerplate/internal/infrastructure/repository"
	ihttp "github.com/pot-code/go-boilerplate/internal/interfaces/http"
	"github.com/pot-code/go-boilerplate/internal/usecase"
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

	UUIDGenerator := infra.NewNanoIDGenerator(option.Security.IDLength)
	UserRepo := auth.NewUserRepository(dbConn, UUIDGenerator)
	UserUserCase := auth.NewUserUseCase(UserRepo)

	LessonRepo := repo.NewLessonRepository(dbConn)
	LessonUseCase := usecase.NewLessonUseCase(LessonRepo)

	TimeSpentRepo := repo.NewTimeSpentRepository(dbConn)
	TimeSpentUseCase := usecase.NewTimeSpentUseCase(TimeSpentRepo)

	ihttp.Serve(option, UserUserCase, UserRepo, LessonUseCase, TimeSpentUseCase)
}
