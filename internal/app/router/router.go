package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sinfirst/URL-Cutter/internal/app/app"
)

func NewRouter(a *app.App) *gin.Engine {
	server := gin.Default()
	server.HandleMethodNotAllowed = true
	server.POST(`/`, a.PostHandler)
	server.GET(`/:id`, a.GetHandler)

	return server
}
