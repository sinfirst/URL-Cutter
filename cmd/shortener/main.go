package main

import (
	"net/http"

	"github.com/sinfirst/URL-Cutter/internal/app/app"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/middleware/logging"
	"github.com/sinfirst/URL-Cutter/internal/app/pg/postgresbd"
	"github.com/sinfirst/URL-Cutter/internal/app/router"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

func main() {
	logger := logging.NewLogger()
	conf := config.NewConfig()
	strg := storage.NewStorage(conf, logger)
	pg := postgresbd.NewPGDB(conf, logger)
	a := app.NewApp(strg, conf, pg, logger)
	router := router.NewRouter(*a)

	logger.Infow("Starting server", "addr", conf.ServerAdress)
	err := http.ListenAndServe(conf.ServerAdress, router)

	if err != nil {
		logger.Fatalw("Can't run server ", err)
	}
}
