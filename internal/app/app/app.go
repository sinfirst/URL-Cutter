package app

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/middleware/jwtgen"
	"github.com/sinfirst/URL-Cutter/internal/app/models"
)

type App struct {
	storage  models.Storage
	config   config.Config
	logger   zap.SugaredLogger
	deleteCh chan string
}

func NewApp(storage models.Storage, config config.Config, logger zap.SugaredLogger, deleteCh chan string) *App {
	app := &App{storage: storage, config: config, logger: logger, deleteCh: deleteCh}
	return app
}

func (a *App) BatchShortenURL(w http.ResponseWriter, r *http.Request) {
	var requests []models.ShortenRequestForBatch
	err := json.NewDecoder(r.Body).Decode(&requests)

	if err != nil {
		a.logger.Errorw("Bad JSON data")
		return
	}

	if len(requests) == 0 {
		a.logger.Errorw("Batch cannot be empty")
		return
	}

	var responces []models.ShortenResponceForBatch
	for _, req := range requests {
		shortURL := fmt.Sprintf("%x", md5.Sum([]byte(req.OriginalURL)))[:8]
		responces = append(responces, models.ShortenResponceForBatch{
			CorrelationID: req.CorrelationID,
			ShortURL:      a.config.Host + "/" + shortURL,
		})
		err = a.storage.SetURL(r.Context(), shortURL, req.OriginalURL, 0)
		if err != nil {
			a.logger.Errorw("Problem with set in storage", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responces)
}

func (a *App) GetHandler(w http.ResponseWriter, r *http.Request) {
	idGet := chi.URLParam(r, "id")
	if origURL, err := a.storage.GetURL(r.Context(), idGet); err == nil {
		w.Header().Set("Location", origURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "URL not found", http.StatusGone)
	}
}
func (a *App) PostHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")

	if err != nil {
		fmt.Print("No token value!")
	}

	UserID := jwtgen.GetUserID(cookie.Value)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	if len(body) == 0 {
		http.Error(w, "url param required", http.StatusBadRequest)
		return
	}

	shortURL := fmt.Sprintf("%x", md5.Sum(body))[:8]
	if _, err := a.storage.GetURL(r.Context(), shortURL); err == nil {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "%s/%s", a.config.Host, shortURL)
		return
	}
	err = a.storage.SetURL(r.Context(), shortURL, string(body), UserID)

	if err != nil {
		a.logger.Errorw("Problem with set in storage", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", a.config.Host, shortURL)
}

func (a *App) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	var input models.OriginalURL
	var output models.ResultURL

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		a.logger.Errorw("Bad JSON OriginalURL")
		return
	}
	shortURL := fmt.Sprintf("%x", md5.Sum([]byte(input.URL)))[:8]
	output = models.ResultURL{Result: a.config.Host + "/" + shortURL}
	JSONResponse, err := json.Marshal(output)
	if err != nil {
		a.logger.Errorw("Problem with create JSONResponse")
		return
	}
	if _, err := a.storage.GetURL(r.Context(), shortURL); err == nil {
		a.logger.Infow("Original URL already in storage")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(JSONResponse)
		return
	}
	err = a.storage.SetURL(r.Context(), shortURL, string(input.URL), 0)
	if err != nil {
		a.logger.Errorw("Problem with set in storage", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(JSONResponse)
}

func (a *App) DBPing(w http.ResponseWriter, r *http.Request) {
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

func (a *App) GetUserUrls(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	var UserID int
	var ShorigURLs []models.ShortenOrigURLs

	if err != nil {
		token, _ := jwtgen.BuildJWTString()
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		fmt.Println("Token created at GetUserUrls!")
		return
	}
	fmt.Println(cookie.Value)
	if err := cookie.Valid(); err == nil {
		UserID = jwtgen.GetUserID(cookie.Value)
		fmt.Println(UserID)
		fmt.Println("UserID collected from cookie.Value")
	}

	URLs, err := a.storage.GetByUserID(r.Context(), UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for key, value := range URLs {
		clear(ShorigURLs)
		ShorigURLs = append(ShorigURLs, models.ShortenOrigURLs{
			ShortURL:    a.config.Host + "/" + key,
			OriginalURL: value,
		})
	}
	fmt.Println(URLs)
	fmt.Println(ShorigURLs)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(ShorigURLs)
	if err != nil {
		panic(err)
	}

}

func (a *App) DeleteUrls(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var urlIDs []string

	err = json.Unmarshal(body, &urlIDs)
	if err != nil {
		http.Error(w, "Ошибка парсинга запроса", http.StatusBadRequest)
		return
	}

	for _, id := range urlIDs {
		a.AddToChan(id)
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *App) CloseCh() {
	close(a.deleteCh)
}

func (a *App) AddToChan(id string) {
	a.deleteCh <- id
}
