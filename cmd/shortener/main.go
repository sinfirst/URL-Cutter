package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sinfirst/URL-Cutter/internal/app/app"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/middleware/logging"
	"github.com/sinfirst/URL-Cutter/internal/app/pg/postgresbd"
	"github.com/sinfirst/URL-Cutter/internal/app/router"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
	"github.com/sinfirst/URL-Cutter/internal/app/workers"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	DeleteCh := make(chan string, 6)
	logger := logging.NewLogger()
	conf := config.NewConfig()
	db := postgresbd.NewPGDB(conf, logger)
	strg := storage.NewStorage(conf, logger)
	a := app.NewApp(strg, conf, logger, DeleteCh)
	router := router.NewRouter(*a)
	workers := workers.NewDeleteWorker(ctx, db, DeleteCh, *a)

	if conf.DatabaseDsn != "" {
		postgresbd.InitMigrations(conf, logger)
	}

	server := &http.Server{Addr: conf.ServerAdress, Handler: router}

	go func() {
		logger.Infow("Starting server", "addr", conf.ServerAdress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalw("create server error: ", err)
		}
	}()

	<-ctx.Done()
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Errorw("Server shutdown error", err)
	}

	workers.StopWorker()
	a.CloseCh()
}
