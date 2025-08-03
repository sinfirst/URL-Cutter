// Package storage пакет с инициализацией хранилища данных
package storage

import (
	"github.com/sinfirst/URL-Cutter/internal/app"
	"github.com/sinfirst/URL-Cutter/internal/config"
	"github.com/sinfirst/URL-Cutter/internal/storage/files"
	"github.com/sinfirst/URL-Cutter/internal/storage/memory"
	"github.com/sinfirst/URL-Cutter/internal/storage/pg/postgresbd"
	"go.uber.org/zap"
)

// NewStorage инициализация storage интерфеса для хранилища данных
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
