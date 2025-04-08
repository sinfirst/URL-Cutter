package app

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/files"
	"github.com/sinfirst/URL-Cutter/internal/app/pg/postgresbd"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

type App struct {
	storage storage.Storage
	config  config.Config
	file    files.File
	pg      *postgresbd.PGDB
	logger  zap.SugaredLogger
}

func NewApp(storage storage.Storage, config config.Config, file files.File, pg *postgresbd.PGDB, logger zap.SugaredLogger) *App {
	app := &App{storage: storage, config: config, file: file, pg: pg, logger: logger}
	return app
}

func (a *App) BatchShortenURL(w http.ResponseWriter, r *http.Request) {
	var requests []storage.ShortenRequestForBatch
	err := json.NewDecoder(r.Body).Decode(&requests)

	if err != nil {
		a.logger.Errorw("Bad JSON data")
		return
	}

	if len(requests) == 0 {
		a.logger.Errorw("Batch cannot be empty")
		return
	}

	var responces []storage.ShortenResponceForBatch
	for _, req := range requests {
		shortURL := fmt.Sprintf("%x", md5.Sum([]byte(req.OriginalURL)))[:8]
		responces = append(responces, storage.ShortenResponceForBatch{
			CorrelationID: req.CorrelationID,
			ShortURL:      a.config.Host + "/" + shortURL,
		})

		a.storage.Set(shortURL, req.OriginalURL)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responces)
}

func (a *App) GetHandler(w http.ResponseWriter, r *http.Request) {
	idGet := chi.URLParam(r, "id")
	if origURL, flag := a.storage.Get(idGet); flag {
		w.Header().Set("Location", origURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		a.logger.Infow("Can't find shortURL in storage")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (a *App) PostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.logger.Errorw("Problem with read original URL")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(string(body)) == 0 {
		a.logger.Errorw("Original URL is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := fmt.Sprintf("%x", md5.Sum(body))[:8]
	if _, flag := a.storage.Get(shortURL); flag {
		a.logger.Infow("Original URL already in storage")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "%s/%s", a.config.Host, shortURL)
		return
	}
	a.storage.Set(shortURL, string(body))
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", a.config.Host, shortURL)
}

func (a *App) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	var input storage.OriginalURL
	var output storage.ResultURL

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		a.logger.Errorw("Bad JSON OriginalURL")
		return
	}
	shortURL := fmt.Sprintf("%x", md5.Sum([]byte(input.URL)))[:8]
	output = storage.ResultURL{Result: a.config.Host + "/" + shortURL}
	JSONResponse, err := json.Marshal(output)
	if err != nil {
		a.logger.Errorw("Problem with create JSONResponse")
		return
	}
	if _, flag := a.storage.Get(shortURL); flag {
		a.logger.Infow("Original URL already in storage")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(JSONResponse)
		return
	}
	a.storage.Set(shortURL, string(input.URL))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(JSONResponse)
}

func (a *App) DBPing(w http.ResponseWriter, r *http.Request) {
	db, _ := a.pg.ConnectToDB()
	defer db.Close()
	w.WriteHeader(http.StatusOK)
}
