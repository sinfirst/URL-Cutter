package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/app"
	"github.com/sinfirst/URL-Cutter/middleware/compress"
	"github.com/sinfirst/URL-Cutter/middleware/logging"
)

func NewRouter(a app.App) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logging.WithLogging)
	router.With(compress.CompressHandle).Post("/", a.PostHandler)
	router.With(compress.CompressHandle).Post("/api/shorten", a.JSONPostHandler)
	router.With(compress.DecompressHandle).Get("/{id}", a.GetHandler)
	return router
}
