package storage

import (
	"github.com/sinfirst/URL-Cutter/internal/app/app"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/storage/files"
	"github.com/sinfirst/URL-Cutter/internal/app/storage/memory"
	"github.com/sinfirst/URL-Cutter/internal/app/storage/pg/postgresbd"
	"go.uber.org/zap"
)

func NewStorage(conf config.Config, logger zap.SugaredLogger) app.Storage {
	if conf.DatabaseDsn != "" {
		logger.Infow("DB config")
		return postgresbd.NewPGDB(conf, logger)
	}
	if conf.FilePath != "" {
		logger.Infow("file config")
		return files.NewFile(conf, logger)
	}
	logger.Infow("memory config")
	return memory.NewMapStorage()
}
