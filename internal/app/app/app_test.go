package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

func TestGet(t *testing.T) {
	stg := storage.NewStorage()
	stg.Set("abcdefgh", "http://mail.ru/")
	cfg := config.NewConfig()
	a := NewApp(stg, cfg)

	testRequest := func(shortURL string) *http.Request {
		req := httptest.NewRequest("GET", "/"+shortURL, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", shortURL)
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		return req.WithContext(ctx)
	}
	tests := []struct {
		name           string
		shortURL       string
		expectedCode   int
		expectedHeader string
	}{
		{
			name:           "valid short URL",
			shortURL:       "abcdefgh",
			expectedCode:   http.StatusTemporaryRedirect,
			expectedHeader: "http://mail.ru/",
		},
		{
			name:           "Invalid req",
			shortURL:       "asajkdashgkj",
			expectedCode:   http.StatusBadRequest,
			expectedHeader: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testRequest(tt.shortURL)
			resRec := httptest.NewRecorder()
			a.GetHandler(resRec, req)

			if resRec.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, resRec.Code)
			}

			locationHeader := resRec.Header().Get("Location")
			if locationHeader != tt.expectedHeader {
				t.Errorf("expected header %v got %v", tt.expectedHeader, locationHeader)
			}
		})
	}
}

/*func TestPost(t *testing.T) {
	stg := storage.NewStorage()
	cfg := config.NewConfig()
	a := NewApp(stg, cfg)

	tests := []struct {
		name         string
		origURL      string
		expectedCode int
	}{
		{
			name:         "valid short URL",
			origURL:      "http://mail.ru/",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Invalid req",
			origURL:      "",
			expectedCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(tt.origURL)))
			resRec := httptest.NewRecorder()
			a.PostHandler(resRec, req)

			if resRec.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, resRec.Code)
			}
		})
	}
}*/
