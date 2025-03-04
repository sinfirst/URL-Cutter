package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

func TestGet(t *testing.T) {
	stg := storage.NewStorage()
	cfg := config.NewConfig()
	a := NewApp(stg, cfg)

	tests := []struct {
		name           string
		method         string
		shortURL       string
		expectedCode   int
		expectedHeader string
	}{
		{
			name:           "Invalid short URL",
			method:         http.MethodGet,
			shortURL:       "asdoif",
			expectedCode:   http.StatusBadRequest,
			expectedHeader: "",
		},
		{
			name:           "Invalid method",
			method:         http.MethodPost,
			shortURL:       "asajkdashgkj",
			expectedCode:   http.StatusBadRequest,
			expectedHeader: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/"+tt.shortURL, nil)
			resRec := httptest.NewRecorder()

			a.PostHandler(resRec, req)

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

func TestPost(t *testing.T) {
	stg := storage.NewStorage()
	cfg := config.NewConfig()
	a := NewApp(stg, cfg)
	tests := []struct {
		name         string
		method       string
		url          string
		expectedCode int
	}{
		{
			name:         "Simple POST request",
			method:       http.MethodPost,
			url:          "https://yandex.ru",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Invalid request",
			method:       http.MethodGet,
			url:          "",
			expectedCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPost {
				req = httptest.NewRequest(tt.method, "http://localhost:8080/", bytes.NewBufferString(tt.url))
			} else {
				req = httptest.NewRequest(tt.method, "http://localhost:8080/", nil)
			}
			rr := httptest.NewRecorder()
			a.PostHandler(rr, req)
			if rr.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rr.Code)
			}
		})
	}
}
