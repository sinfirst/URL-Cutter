package router

import (
	"net/http"

	"github.com/sinfirst/URL-Cutter/internal/app/app"
)

func NewRouter(a *app.App) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("POST /", a.PostHandler)
	router.HandleFunc("GET /{id}", a.GetHandler)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	return router
}
