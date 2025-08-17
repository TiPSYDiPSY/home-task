package operation

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Init(r *chi.Mux) {
	r.Use(middleware.Heartbeat("/ping"))
}
