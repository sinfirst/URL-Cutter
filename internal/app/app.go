// Package app пакет со всеми эндпоинтами сервиса
package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/handlers"
	"github.com/sinfirst/URL-Cutter/internal/middleware/jwtgen"
	"github.com/sinfirst/URL-Cutter/internal/models"
)

// App структура для хранения переменных
type App struct {
	logger  zap.SugaredLogger
	handler handlers.Handler
}

// NewApp конструктор для App
func NewApp(logger zap.SugaredLogger, handler handlers.Handler) *App {
	app := &App{logger: logger, handler: handler}
	return app
}

// BatchShortenURL функция для добавления группы урлов в бд
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
	resp, err := a.handler.BatchShortenURL(r.Context(), requests)
	if err != nil {
		a.logger.Errorw("Problem with set in storage", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetHandler осуществляет редирект на полный урл, если короткий урл есть в базе
func (a *App) GetHandler(w http.ResponseWriter, r *http.Request) {
	idGet := chi.URLParam(r, "id")
	if origURL, err := a.handler.GetHandler(r.Context(), idGet); err == nil {
		w.Header().Set("Location", origURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "URL not found", http.StatusGone)
	}
}

// PostHandler осуществляет сокращение урла, переданного с помощью text/plain
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
	shortURL, typeOfErr, err := a.handler.PostHandler(r.Context(), body, UserID)
	if typeOfErr == 1 {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(shortURL))
		return
	} else if typeOfErr == 2 {
		a.logger.Errorw("problem with set in storage", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// JSONPostHandler осуществляет сокращение урла, переданного с помощью JSON
func (a *App) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	var input models.OriginalURL

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		a.logger.Errorw("Bad JSON OriginalURL")
		return
	}

	JSONResponse, typeOfError, err := a.handler.JSONPostHandler(r.Context(), input)
	if typeOfError == 1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(JSONResponse)
	} else if typeOfError == 2 {
		a.logger.Errorw("problem with marshal json", err)
	} else if typeOfError == 3 {
		a.logger.Errorw("problem with set in storage", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(JSONResponse)
}

// DBPing осуществляет ping до базы данных
func (a *App) DBPing(w http.ResponseWriter, r *http.Request) {
	err := a.handler.DBPing()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetUserUrls по этому хендлеру получаем все урлы конкретного пользователя
func (a *App) GetUserUrls(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	var userID int

	if err != nil {
		token, err := jwtgen.BuildJWTString()
		if err != nil {
			a.logger.Errorw("can't make jwt token")
			return
		}
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
		userID = jwtgen.GetUserID(cookie.Value)
		fmt.Println(userID)
		fmt.Println("UserID collected from cookie.Value")
	}
	urls, err := a.handler.GetUserUrls(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(urls)
	if err != nil {
		a.logger.Errorf("can't encode json", err)
		return
	}

}

// DeleteUrls удаляет запрошенные урлы
func (a *App) DeleteUrls(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = a.handler.DeleteUrls(body)
	if err != nil {
		a.logger.Errorw("Ошибка парсинга json")
	}
	w.WriteHeader(http.StatusAccepted)
}

// GetStats получает статистику сервера(кол-во сокращенных urlов и кол-во уникальных пользователей)
func (a *App) GetStats(w http.ResponseWriter, r *http.Request) {
	ipFromRequest := r.Header.Get("X-Real-IP")
	resp, typeOfError, err := a.handler.GetStats(r.Context(), ipFromRequest)
	if typeOfError == 1 {
		w.WriteHeader(http.StatusForbidden)
		return
	} else if typeOfError == 2 {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}
}
