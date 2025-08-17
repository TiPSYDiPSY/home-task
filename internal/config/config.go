package config

import (
	"fmt"

	"github.com/TiPSYDiPSY/home-task/internal/util/env"
)

type PostgresDBConfig struct {
	Username string
	Password string
	Database string
	Host     string
	Port     int
	SSLMode  string
}

type ServerConfig struct {
	Port                      string
	DatabaseConnectionDetails PostgresDBConfig
}

func NewServerConfig() *ServerConfig {
	config := &ServerConfig{
		Port: env.GetEnv("PORT", "8080"),
		DatabaseConnectionDetails: PostgresDBConfig{
			Username: env.GetEnv("POSTGRES_USER", "myuser"),
			Password: env.GetEnv("POSTGRES_PASS", "mypassword"),
			Database: env.GetEnv("POSTGRES_DB", "mydb"),
			Host:     env.GetEnv("POSTGRES_HOST", "localhost"),
			Port:     env.GetEnvInt("POSTGRES_PORT", "5432"),
		},
	}

	return config
}

func (c PostgresDBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.getSSLMode())
}

func (c PostgresDBConfig) getSSLMode() string {
	if c.SSLMode == "" {
		return "disable"
	}

	return c.SSLMode
}
