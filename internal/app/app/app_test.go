package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

func TestHanedlers(t *testing.T) {
	stg := storage.NewStorage()
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
		name             string
		origURL          string
		expectedPostCode int
		expectedGetCode  int
		expectedHeader   string
		expectedBody     string
	}{
		{
			name:             "test #1",
			origURL:          "http://mail.ru/",
			expectedGetCode:  http.StatusTemporaryRedirect,
			expectedPostCode: http.StatusCreated,
			expectedHeader:   "http://mail.ru/",
		},
		{
			name:             "Invalid req",
			origURL:          "",
			expectedPostCode: http.StatusBadRequest,
			expectedHeader:   "",
			expectedBody:     "url param required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", bytes.NewBufferString((tt.origURL)))
			resRec := httptest.NewRecorder()
			a.PostHandler(resRec, req)
			if resRec.Code != tt.expectedPostCode {
				t.Errorf("expected status %d, got %d", tt.expectedPostCode, resRec.Code)
			}
			if tt.name == "Invalid req" {
				return
			}
			body := resRec.Body.String()
			shortURL := strings.Split(body, "//")
			shortURL = strings.Split(shortURL[1], "/")
			req = testRequest(shortURL[len(shortURL)-1])
			resRec = httptest.NewRecorder()
			a.GetHandler(resRec, req)

			if resRec.Code != tt.expectedGetCode {
				t.Errorf("expected status %d, got %d", tt.expectedGetCode, resRec.Code)
			}

			locationHeader := resRec.Header().Get("Location")
			if locationHeader != tt.expectedHeader {
				t.Errorf("expected header %v got %v", tt.expectedHeader, locationHeader)
			}
		})
	}
}
