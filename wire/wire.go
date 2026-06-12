//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/Perlishnov/TODO_APP/internal/config"
	"github.com/Perlishnov/TODO_APP/internal/controller"
	"github.com/Perlishnov/TODO_APP/internal/dao"
	"github.com/Perlishnov/TODO_APP/internal/middleware"
	"github.com/Perlishnov/TODO_APP/internal/service"
	"github.com/Perlishnov/TODO_APP/internal/utils"
	"github.com/Perlishnov/TODO_APP/pkg/database"
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InitApp() (*App, error) {
	wire.Build(
		config.Load,
		provideLogger,
		provideMongo,
		provideJWTService,
		dao.NewUserDAO,
		dao.NewTaskDAOMongo,
		service.NewAuthService,
		service.NewTaskService,
		controller.NewAuthController,
		controller.NewTaskController,
		middleware.NewAuthMiddleware,
		newApp,
	)
	return nil, nil
}

func provideLogger(cfg *config.Config) *logrus.Logger {
	return utils.NewLogger(cfg.LogLevel)
}

func provideMongo(cfg *config.Config, logger *logrus.Logger) (*mongo.Database, error) {
	return database.NewMongoConnection(cfg, logger)
}

func provideJWTService(cfg *config.Config) utils.JWTService {
	return utils.NewJWTUtil(cfg)
}

type App struct {
	AuthController  *controller.AuthController
	TaskController  *controller.TaskController
	AuthMiddleware  *middleware.AuthMiddleware
	Logger          *logrus.Logger
	DB              *mongo.Database
}

func newApp(
	authCtrl *controller.AuthController,
	taskCtrl *controller.TaskController,
	authMW *middleware.AuthMiddleware,
	logger *logrus.Logger,
	db *mongo.Database,
) *App {
	return &App{
		AuthController: authCtrl,
		TaskController: taskCtrl,
		AuthMiddleware: authMW,
		Logger:         logger,
		DB:             db,
	}
}