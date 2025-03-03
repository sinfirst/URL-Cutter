package app

// import (
// 	"bytes"
// 	"example/internal/app/router"
// 	"example/internal/app/storage"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestTestHandler(t *testing.T) {
// 	tests := []struct {
// 		name             string
// 		URL              string
// 		wantedPostStatus int
// 		wantedGetStatus  int
// 	}{
// 		{
// 			name:             "simple test #1",
// 			URL:              "http://example.ru/",
// 			wantedPostStatus: http.StatusCreated,
// 			wantedGetStatus:  http.StatusTemporaryRedirect,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			stg := storage.NewStorage()
// 			a := App{storage: stg}
// 			rout := router.NewRouter(&a)
// 			req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(tt.URL)))
// 			rr := httptest.NewRecorder()
// 			a.PostHandler(rr, req)
// 			if rr.Code != tt.wantedPostStatus {
// 				t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
// 			}
// 			shortURL := rr.Body.String()[22:]
// 			fmt.Println("test 1 ", shortURL)
// 			req = httptest.NewRequest("GET", "/"+shortURL, nil)
// 			fmt.Println(req.URL.Path)

// 			rr = httptest.NewRecorder()
// 			fmt.Println("test 2 ", a.storage)
// 			a.GetHandler(rr, req)
// 			if rr.Code != tt.wantedGetStatus {
// 				t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rr.Code)
// 			}
// 			if loc := rr.Header().Get("Location"); loc != tt.URL {
// 				t.Errorf("expected Location header %s, got %s", tt.URL, loc)
// 			}
// 		})

// 	}
// }
