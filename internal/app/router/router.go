package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/app"
)

func NewRouter(a app.App) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/", a.PostHandler)
	router.Get("/{id}", a.GetHandler)
	return router
}
