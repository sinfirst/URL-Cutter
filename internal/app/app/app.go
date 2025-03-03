package app

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

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
	idGet := r.PathValue("id")
	fmt.Println("test 3 ", idGet)
	if origURL, flag := a.storage.Get(idGet); flag {
		w.Header().Set("Location", origURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "URL not found", http.StatusBadRequest)
	}
}

func (a *App) PostHandler(w http.ResponseWriter, r *http.Request) {
	var shortURL string
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, "url is empty", http.StatusBadRequest)
		return
	}
	for {
		shortURL = a.getShortURL()
		if _, flag := a.storage.Get(shortURL); !flag {
			a.storage.Set(shortURL, string(body))
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "%s/%s", a.config.ServerAdress, shortURL)
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
