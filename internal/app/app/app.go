package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/files"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

type App struct {
	storage storage.Storage
	config  config.Config
	file    *files.File
}

func NewApp(storage *storage.MapStorage, config config.Config, file *files.File) *App {
	return &App{storage: storage, config: config, file: file}
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

	if len(string(body)) == 0 {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	for {
		shortURL = a.getShortURL()
		if _, flag := a.storage.Get(shortURL); !flag {
			a.storage.Set(shortURL, string(body))
			a.file.UpdateFile(files.JSONStructForBD{
				ShortURL:    shortURL,
				OriginalURL: string(body),
			})
			a.addDataToDB(shortURL, string(body))
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "%s/%s", a.config.Host, shortURL)
			break
		}
	}
}

func (a *App) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	var shortURL string
	var input storage.OriginalURL
	var output storage.ResultURL
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body := input.URL
	for {
		shortURL = a.getShortURL()
		if _, flag := a.storage.Get(shortURL); !flag {
			a.storage.Set(shortURL, string(body))
			a.file.UpdateFile(files.JSONStructForBD{
				ShortURL:    shortURL,
				OriginalURL: string(body),
			})
			a.addDataToDB(shortURL, string(body))
			output = storage.ResultURL{Result: a.config.Host + "/" + shortURL}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			resp, err := json.Marshal(output)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(resp)
			break
		}
	}
}

func (a *App) CheckConnectToDB(w http.ResponseWriter, _ *http.Request) {
	db, err := sql.Open("pgx", a.config.DatabaseDsn)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	err = db.Ping()

	if err != nil {
		http.Error(w, "Failed ping to database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func (a *App) addDataToDB(shortURL, originalURL string) {
	db, err := sql.Open("pgx", a.config.DatabaseDsn)
	if err != nil {
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS urls (
		short_url TEXT NOT NULL PRIMARY KEY,
		original_url TEXT NOT NULL
	);`)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`INSERT INTO urls VALUES ($1, $2)`, shortURL, originalURL)
	if err != nil {
		panic(err)
	}
}

func (a *App) getShortURL() string {
	var res string
	for i := 0; i < 8; i++ {
		res += a.config.Letters[rand.Intn(len(a.config.Letters))]
	}
	return res
}
