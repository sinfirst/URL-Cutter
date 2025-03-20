package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/app"
	"github.com/sinfirst/URL-Cutter/middleware/logging"
	"go.uber.org/zap"
)

func NewRouter(a app.App, sugar zap.SugaredLogger) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/", logging.WithLogging(a.PostHandler, sugar))
	router.Post("/api/shorten/", logging.WithLogging(a.JSONPostHandler, sugar))
	router.Get("/{id}", logging.WithLogging(a.GetHandler, sugar))
	return router
}
