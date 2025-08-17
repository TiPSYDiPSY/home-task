package public

import (
	"context"

	"github.com/TiPSYDiPSY/home-task/internal/api/handler/public/handlers/middleware"

	"github.com/TiPSYDiPSY/home-task/internal/api/handler/public/handlers/user"

	"github.com/go-chi/chi/v5"

	"github.com/TiPSYDiPSY/home-task/internal/config"
	"github.com/TiPSYDiPSY/home-task/internal/service"
)

func Init(ctx context.Context, c *config.ServerConfig, container service.Container, mainRouter *chi.Mux) {
	subRouter := chi.NewRouter()

	subRouter.Group(func(r chi.Router) {
		r.Use(middleware.SourceTypeValidator())
		r.Post("/{userID}/transaction", user.UpdateBalance(container.UserService))
	})

	subRouter.Group(func(r chi.Router) {
		r.Get("/{userID}/balance", user.GetBalance(container.UserService))
	})

	mainRouter.Mount("/user", subRouter)
}
