package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

func TestGet(t *testing.T) {
	stg := storage.NewStorage()
	cfg := config.NewConfig()
	a := NewApp(stg, cfg)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/:id", a.GetHandler)
	a.storage.Set("abcdefgh", "http://mail.ru/")

	tests := []struct {
		name           string
		shortURL       string
		expectedCode   int
		expectedHeader string
		expectedBody   string
	}{
		{
			name:           "valid short URL",
			shortURL:       "abcdefgh",
			expectedCode:   http.StatusTemporaryRedirect,
			expectedHeader: "http://mail.ru/",
			expectedBody:   "{}",
		},
		{
			name:           "Invalid method",
			shortURL:       "asajkdashgkj",
			expectedCode:   http.StatusBadRequest,
			expectedHeader: "",
			expectedBody:   "{}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/"+tt.shortURL, nil)
			resRec := httptest.NewRecorder()

			r.ServeHTTP(resRec, req)

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
