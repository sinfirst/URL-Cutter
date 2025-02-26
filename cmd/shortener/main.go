package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var letters = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", "")

type DataBase struct {
	mainURL  []string
	shortURL []string
	countURL int
}

var base = DataBase{}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", webhook)
	mux.HandleFunc("/{id}", GetHandler)
	return http.ListenAndServe(":8080", mux)
}

func webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostHandler(w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	idGet := r.PathValue("id")
	flag, idURL := checkInBdShortURL(idGet)
	if flag {
		w.Header().Set("Location", base.mainURL[idURL])
		w.WriteHeader(http.StatusTemporaryRedirect)

	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	var url, shortURL string
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	url = string(body)
	flag, id := checkInBdMainURL(url)
	if flag {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("http://localhost:8080/" + base.shortURL[id]))
		return
	}
	shortURL = getShortURL()
	for {
		flag, _ = checkInBdMainURL(shortURL)
		if flag {
			shortURL = getShortURL()
		} else {
			break
		}
	}
	base.mainURL = append(base.mainURL, url)
	base.shortURL = append(base.shortURL, shortURL)
	base.countURL += 1
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("http://localhost:8080/" + shortURL))
}
func getShortURL() string {
	var res string
	for i := 0; i < 8; i++ {
		res += letters[rand.Intn(len(letters))]
	}
	return res
}

func checkInBdShortURL(url string) (bool, int) {
	for i := 0; i < base.countURL; i++ {
		if base.shortURL[i] == url {
			return true, i
		}
	}
	return false, -1
}

func checkInBdMainURL(url string) (bool, int) {
	for i := 0; i < base.countURL; i++ {
		if base.mainURL[i] == url {
			return true, i
		}
	}
	return false, -1
}
