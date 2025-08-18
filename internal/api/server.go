package api

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TiPSYDiPSY/home-task/internal/api/handler/operation"
	"github.com/TiPSYDiPSY/home-task/internal/api/handler/public"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/TiPSYDiPSY/home-task/internal/config"
	"github.com/TiPSYDiPSY/home-task/internal/service"
)

const (
	readTimeoutSec     = 5
	writeTimeoutSec    = 60
	idleTimeoutSec     = 120
	shutdownTimeoutSec = 30
)

func StartServer(ctx context.Context, c *config.ServerConfig, container service.Container) {
	log := logrus.WithContext(ctx)
	log.Info("Starting http server on port: " + c.Port)

	router := initServerMux(container)

	srv := &http.Server{
		ReadTimeout:  readTimeoutSec * time.Second,
		WriteTimeout: writeTimeoutSec * time.Second,
		IdleTimeout:  idleTimeoutSec * time.Second,
		Addr:         ":" + c.Port,
		Handler:      router,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Fatal("Server failed to start")
		}
	}()

	log.Info("Server started successfully")

	<-quit
	log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeoutSec*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")

		return
	}

	log.Info("Server exited gracefully")
}

func initServerMux(container service.Container) *chi.Mux {
	r := chi.NewRouter()

	operation.Init(r)
	public.Init(container, r)

	return r
}
