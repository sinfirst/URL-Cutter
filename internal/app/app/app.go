package app

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

type App struct {
	storage storage.Storage
	config  config.Config
}

func NewApp(storage storage.Storage, config config.Config) *App {
	return &App{storage: storage, config: config}
}

func (a *App) GetHandler(w http.ResponseWriter, r *http.Request) {
	idGet := chi.URLParam(r, "id")
	if origURL, flag := a.storage.Get(idGet); flag {
		w.Header().Set("Location", origURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (a *App) PostHandler(w http.ResponseWriter, r *http.Request) {
	var shortURL string
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	if len(body) == 0 {
		http.Error(w, "url param required", http.StatusBadRequest)
		return
	}
	_, err = url.ParseRequestURI(string(body))

	if err != nil {
		http.Error(w, "Correct url required", http.StatusBadRequest)
	}
	for {
		shortURL = a.getShortURL()
		if _, flag := a.storage.Get(shortURL); flag {
			a.storage.Set(shortURL, string(body))
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "%s/%s", a.config.Host, shortURL)
			break
		}
	}
}

func (a *App) getShortURL() string {
	var res string
	for i := 0; i < 8; i++ {
		res += a.config.Letters[rand.Intn(len(a.config.Letters))]
	}
	return res
}
