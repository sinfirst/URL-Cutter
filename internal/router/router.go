// Package router пакет с инициализацией маршутизации
package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app"
	"github.com/sinfirst/URL-Cutter/internal/middleware/compress"
	"github.com/sinfirst/URL-Cutter/internal/middleware/jwtgen"
	"github.com/sinfirst/URL-Cutter/internal/middleware/logging"
)

// NewRouter описание всех эндпоинтов
func NewRouter(a *app.App) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logging.WithLogging)
	router.With(compress.DecompressHandle, jwtgen.AuthMiddleware).Post("/", a.PostHandler)
	router.With(compress.DecompressHandle).Post("/api/shorten", a.JSONPostHandler)
	router.With(compress.DecompressHandle).Post("/api/shorten/batch", a.BatchShortenURL)
	router.With(compress.CompressHandle).Get("/{id}", a.GetHandler)
	router.With(compress.CompressHandle).Get("/api/internal/stats", a.GetStats)

	router.Get("/ping", a.DBPing)
	router.Get("/api/user/urls", a.GetUserUrls)
	router.Delete("/api/user/urls", a.DeleteUrls)

	return router
}
