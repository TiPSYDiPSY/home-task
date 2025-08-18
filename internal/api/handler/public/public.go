package public

import (
	"github.com/TiPSYDiPSY/home-task/internal/api/handler/public/handlers/middleware"
	"github.com/TiPSYDiPSY/home-task/internal/api/handler/public/handlers/user"
	"github.com/TiPSYDiPSY/home-task/internal/util/validation"

	"github.com/go-chi/chi/v5"

	"github.com/TiPSYDiPSY/home-task/internal/service"
)

func Init(container service.Container, mainRouter *chi.Mux) {
	subRouter := chi.NewRouter()

	loggingMiddleware := middleware.NewLoggingMiddleware(middleware.LoggingConfig{
		BodyLoggingEnabled: true,
		ServiceName:        "home-task",
	})

	subRouter.Use(loggingMiddleware.Middleware)

	subRouter.Group(func(r chi.Router) {
		r.Use(middleware.SourceTypeValidator)
		r.Post("/{userID}/transaction", user.UpdateBalance(container.UserService, validation.NewValidator()))
	})

	subRouter.Group(func(r chi.Router) {
		r.Get("/{userID}/balance", user.GetBalance(container.UserService))
	})

	mainRouter.Mount("/user", subRouter)
}
