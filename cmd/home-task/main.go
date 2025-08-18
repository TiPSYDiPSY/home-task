package main

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/TiPSYDiPSY/home-task/internal/api"
	"github.com/TiPSYDiPSY/home-task/internal/config"
	"github.com/TiPSYDiPSY/home-task/internal/db"
	"github.com/TiPSYDiPSY/home-task/internal/service"
)

func main() {
	ctx := context.Background()

	shutdown := config.InitTracer()
	defer shutdown()

	servConfig := config.NewServerConfig()
	logger := logrus.WithContext(ctx)

	ds, err := db.NewPostgresDBDataStore(ctx, servConfig.DatabaseConnectionDetails)
	if err != nil {
		logger.WithError(err).Fatal("Connect to DB failed with error")
	}

	if err := ds.RunAutoMigrate(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to run database migrations")
	}

	container := service.NewContainer(ds)

	api.StartServer(ctx, servConfig, container)
}
