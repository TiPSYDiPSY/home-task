package db

import (
	"context"
	"fmt"

	"github.com/TiPSYDiPSY/home-task/internal/db/gorm_logger"

	"github.com/TiPSYDiPSY/home-task/internal/config"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDBDataStore struct {
	db *gorm.DB
}

func NewPostgresDBDataStore(ctx context.Context, c config.PostgresDBConfig) (*PostgresDBDataStore, error) {
	log := logrus.WithContext(ctx)

	log.Info("Connecting to DB...")

	db, err := gorm.Open(postgres.Open(c.DSN()), &gorm.Config{
		Logger: gorm_logger.New(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Successfully connected to DB")

	return &PostgresDBDataStore{db: db}, nil
}
