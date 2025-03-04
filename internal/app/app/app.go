package app

import (
	"io"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

type App struct {
	storage storage.Storage
	config  config.Config
}

func NewApp(storage storage.Storage, config config.Config) *App {
	return &App{storage: storage, config: config}
}

func (a *App) GetHandler(ctx *gin.Context) {
	idGet := ctx.Param("id")
	if origURL, flag := a.storage.Get(idGet); flag {
		ctx.Header("Location", origURL)
		ctx.IndentedJSON(http.StatusTemporaryRedirect, gin.H{})
	} else {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{})
	}
}

func (a *App) PostHandler(ctx *gin.Context) {
	var shortURL string
	url, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err})
	}
	if string(url) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "url param required"})
		return
	}
	for {
		shortURL = a.getShortURL()
		if _, flag := a.storage.Get(shortURL); !flag {
			a.storage.Set(shortURL, string(url))
			ctx.String(http.StatusCreated, "http://localhost:8080/"+shortURL)
			break
		}
	}
}

func (a *App) getShortURL() string {
	var res string
	for i := 0; i < 8; i++ {
		res += a.config.Letters[rand.Intn(len(a.config.Letters))]
	}
	return res
}
