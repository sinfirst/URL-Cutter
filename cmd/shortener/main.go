package main

import (
	"fmt"
	"net/http"

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

	err := http.ListenAndServe(conf.Host, rout)
	if err != nil {
		fmt.Println(err)
	}
}
