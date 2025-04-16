package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/middleware/logging"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

func TestRedirect(t *testing.T) {
	m1 := storage.NewMapStorage()
	m1.SetURL("abc123", "https://example.com")
	logger := logging.NewLogger()
	app := &App{storage: m1, logger: logger}

	testRequest := func(shortURL string) *http.Request {
		req := httptest.NewRequest("GET", "/"+shortURL, nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", shortURL)

		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		return req.WithContext(ctx)
	}

	t.Run("valid short URL redirects", func(t *testing.T) {
		req := testRequest("abc123")
		rr := httptest.NewRecorder()

		app.GetHandler(rr, req)

		if rr.Code != http.StatusTemporaryRedirect {
			t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rr.Code)
		}

		expectedLocation := "https://example.com"
		if loc := rr.Header().Get("Location"); loc != expectedLocation {
			t.Errorf("expected Location header %s, got %s", expectedLocation, loc)
		}
	})

	t.Run("invalid short URL returns 400", func(t *testing.T) {
		req := testRequest("invalid")
		rr := httptest.NewRecorder()

		app.GetHandler(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

	})
}

// func TestShortenURL(t *testing.T) {
// 	m1 := storage.NewStorage()
// 	m1.Set("abc123", "https://example.com")
// 	logger := logging.NewLogger()
// 	conf := config.NewConfig()
// 	file := files.NewFile(conf, m1)
// 	pg := postgresbd.NewPGDB(conf, logger, m1, file)
// 	app := NewApp(m1, conf, file, pg)

// 	t.Run("successful URL shortening", func(t *testing.T) {
// 		originalURL := "https://example.com"
// 		reqBody := []byte(originalURL)

// 		req := httptest.NewRequest("POST", "/", bytes.NewReader(reqBody))
// 		rr := httptest.NewRecorder()

// 		app.PostHandler(rr, req)

// 		if rr.Code != http.StatusCreated {
// 			t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
// 		}

// 	})

// 	t.Run("empty body returns 400", func(t *testing.T) {
// 		req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte{}))
// 		rr := httptest.NewRecorder()

// 		app.PostHandler(rr, req)

// 		if rr.Code != http.StatusBadRequest {
// 			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
// 		}
// 	})
// }

// type RequestPayload struct {
// 	URL string `json:"url"`
// }

// func TestShortenURLJSON(t *testing.T) {
// 	m1 := storage.NewStorage()
// 	m1.Set("abc123", "https://example.com")
// 	logger := logging.NewLogger()
// 	conf := config.NewConfig()
// 	file := files.NewFile(conf, m1)
// 	pg := postgresbd.NewPGDB(conf, logger, m1, file)
// 	app := NewApp(m1, conf, file, pg)
// 	tests := []struct {
// 		name           string
// 		requestBody    string
// 		expectedStatus int
// 		expectedResult string
// 	}{
// 		{
// 			name:           "Valid URL",
// 			requestBody:    `{"url":"https://example.com"}`,
// 			expectedStatus: http.StatusCreated,
// 			expectedResult: "https://example.com",
// 		},
// 		{
// 			name:           "Invalid JSON",
// 			requestBody:    `{"url":}`,
// 			expectedStatus: http.StatusBadRequest,
// 			expectedResult: "",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			req := httptest.NewRequest(http.MethodPost, "/shorten/api", bytes.NewBufferString(tt.requestBody))
// 			req.Header.Set("Content-Type", "application/json")
// 			w := httptest.NewRecorder()

// 			app.JSONPostHandler(w, req)
// 			resp := w.Result()
// 			defer resp.Body.Close()

// 			if resp.StatusCode != tt.expectedStatus {
// 				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
// 			}

// 			if resp.StatusCode == http.StatusCreated {
// 				var res storage.ResultURL
// 				if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
// 					t.Errorf("failed to decode response: %v", err)
// 				}

// 			}
// 		})
// 	}
// }
