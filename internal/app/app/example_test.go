package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/models"
	"github.com/sinfirst/URL-Cutter/internal/app/storage/memory"
)

func ExampleApp_PostHandler() {
	// Подготовка к выполнению запроса
	m1 := memory.NewMapStorage()
	m1.SetURL(context.Background(), "abc123", "https://example.com", 0)
	conf := config.NewConfig()
	app := &App{
		storage: m1,
		config:  conf,
	}
	// инициализируем body запроса
	originalURL := "https://example.com"
	reqBody := []byte(originalURL)

	req := httptest.NewRequest("POST", "/shorten", bytes.NewReader(reqBody))
	// Добавляем куки
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: "2",
	})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(app.PostHandler)

	// Вызываем хендлер
	handler.ServeHTTP(rr, req)
	resp := rr.Result()

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))

	// Output:
	// 201
	// http://localhost:8080/c984d06a
}

func ExampleApp_JSONPostHandler() {
	// Подготовка к выполнению запроса
	m1 := memory.NewMapStorage()
	m1.SetURL(context.Background(), "abc123", "https://example.com", 0)
	app := &App{
		storage: memory.NewMapStorage(),
		config:  config.Config{Host: "http://localhost"},
	}
	// инициализируем body запроса
	requestBody := `{"url":"https://example.com"}`

	// Подготовка запроса POST /api/shorten
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.JSONPostHandler)

	// Вызываем хендлер
	handler.ServeHTTP(rr, req)
	resp := rr.Result()
	defer resp.Body.Close()

	var res models.ResultURL
	json.NewDecoder(resp.Body).Decode(&res)

	// Выводим код ответа и Body ответа
	fmt.Println(resp.StatusCode)
	fmt.Println(res.Result)

	// Output:
	// 201
	// http://localhost/c984d06a
}

func ExampleApp_GetHandler() {
	// Подготовка к выполнению запроса
	m1 := memory.NewMapStorage()
	m1.SetURL(context.Background(), "c984d06a", "https://example.com", 0)
	app := &App{storage: m1}

	// Используем chi роутер, чтобы брать id запроса
	r := chi.NewRouter()
	r.Get("/{id}", app.GetHandler)

	// Подготовка запроса Get /c984d06a
	req := httptest.NewRequest(http.MethodGet, "/c984d06a", nil)
	rr := httptest.NewRecorder()

	// Вызываем хендлер
	r.ServeHTTP(rr, req)

	// Выводим код ответа и значение заголовка "Location"
	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Location"))

	// Output:
	// 307
	// https://example.com
}
