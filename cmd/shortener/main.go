package main

import (
	"net/http"

	"github.com/sinfirst/URL-Cutter/internal/app/app"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/files"
	"github.com/sinfirst/URL-Cutter/internal/app/router"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
	"github.com/sinfirst/URL-Cutter/middleware/logging"
)

func main() {
	logger := logging.NewLogger()
	conf := config.NewConfig()
	strg := storage.NewStorage()
	file := files.NewFile(conf, strg)
	a := app.NewApp(strg, conf, *file)
	rout := router.NewRouter(*a)

	logger.Infow("Starting server", "addr", conf.ServerAdress)
	http.ListenAndServe(conf.ServerAdress, rout)
}
