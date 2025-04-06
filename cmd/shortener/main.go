package main

import (
	"net/http"

	"github.com/sinfirst/URL-Cutter/internal/app/app"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/files"
	"github.com/sinfirst/URL-Cutter/internal/app/middleware/logging"
	"github.com/sinfirst/URL-Cutter/internal/app/postgresBD"
	"github.com/sinfirst/URL-Cutter/internal/app/router"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

func main() {
	logger := logging.NewLogger()
	conf := config.NewConfig()
	strg := storage.NewStorage()
	file := files.NewFile(conf, strg)
	pg := postgresBD.NewPGDB(conf, logger, strg, file)
	a := app.NewApp(strg, conf, file, pg)
	rout := router.NewRouter(*a, *pg)

	logger.Infow("Starting server", "addr", conf.ServerAdress)
	err := http.ListenAndServe(conf.ServerAdress, rout)
	if err != nil {
		logger.Panicf("Can't run server")
	}
}
