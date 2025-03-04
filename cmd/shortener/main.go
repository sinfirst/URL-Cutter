package main

import (
	"github.com/sinfirst/URL-Cutter/internal/app/app"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/router"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

func main() {
	conf := config.NewConfig()
	strg := storage.NewStorage()
	a := app.NewApp(strg, conf)
	rout := router.NewRouter(a)
	rout.Run(conf.ServerAdress)
}
