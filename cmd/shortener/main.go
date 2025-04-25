package main

import "fmt"

func main() {
	fmt.Println("Test")

	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer cancel()
	// DeleteCh := make(chan string, 6)
	// logger := logging.NewLogger()
	// conf := config.NewConfig()
	// db := postgresbd.NewPGDB(conf, logger)
	// strg := storage.NewStorage(conf, logger)
	// a := app.NewApp(strg, conf, logger, DeleteCh)
	// router := router.NewRouter(*a)
	// workers := workers.NewDeleteWorker(ctx, db, DeleteCh, *a)

	// if conf.DatabaseDsn != "" {
	// 	postgresbd.InitMigrations(conf, logger)
	// }

	// logger.Infow("Starting server", "addr", conf.ServerAdress)
	// err := http.ListenAndServe(conf.ServerAdress, router)

	// if err != nil {
	// 	logger.Fatalw("Can't run server ", err)
	// }
	// workers.StopWorker()
	// a.CloseCh()
}
