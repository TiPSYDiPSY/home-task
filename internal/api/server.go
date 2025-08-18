package api

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/TiPSYDiPSY/home-task/internal/api/handler/operation"
	"github.com/TiPSYDiPSY/home-task/internal/api/handler/public"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/TiPSYDiPSY/home-task/internal/config"
	"github.com/TiPSYDiPSY/home-task/internal/service"
)

const (
	readTimeoutSec  = 5
	writeTimeoutSec = 60
)

func StartServer(ctx context.Context, c *config.ServerConfig, container service.Container) {
	log := logrus.WithContext(ctx)
	log.Info("Starting http server on port: " + c.Port)

	router := initServerMux(container)

	srv := &http.Server{
		ReadTimeout:  readTimeoutSec * time.Second,
		WriteTimeout: writeTimeoutSec * time.Second,
		Addr:         ":" + c.Port,
		Handler:      router,
		// Disable HTTP/2 to force HTTP/1.1 only
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.WithContext(ctx).Fatal(srv.ListenAndServe())
}

func initServerMux(container service.Container) *chi.Mux {
	r := chi.NewRouter()

	operation.Init(r)
	public.Init(container, r)

	return r
}
