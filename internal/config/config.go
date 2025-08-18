package config

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"

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

const (
	TracingSampleRate = 0.1
)

func NewServerConfig() *ServerConfig {
	config := &ServerConfig{
		Port: env.GetEnv("PORT", "8080"),
		DatabaseConnectionDetails: PostgresDBConfig{
			Username: env.GetEnv("DB_USER", "myuser"),
			Password: env.GetEnv("DB_PASSWORD", "mypassword"),
			Database: env.GetEnv("DB_NAME", "mydb"),
			Host:     env.GetEnv("DB_HOST", "localhost"),
			Port:     env.GetEnvInt("DB_PORT", "5432"),
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

func InitTracer() func() {
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(TracingSampleRate)),
	)

	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}
}
