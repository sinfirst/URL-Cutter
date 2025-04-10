package storage

import (
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/files"
	"github.com/sinfirst/URL-Cutter/internal/app/pg/postgresbd"
	"go.uber.org/zap"
)

type OriginalURL struct {
	URL string `json:"url"`
}
type ResultURL struct {
	Result string `json:"result"`
}
type ShortenRequestForBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenResponceForBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
type Storage interface {
	Set(key, value string) bool
	Get(key string) (string, bool)
}

type MapStorage struct {
	data map[string]string
}

func NewMapStorage() *MapStorage {
	return &MapStorage{data: make(map[string]string)}
}

func NewStorage(conf config.Config, logger zap.SugaredLogger) Storage {
	if conf.DatabaseDsn != "" {
		logger.Infow("DB config")
		if conf.DatabaseDsn != "" {
			postgresbd.InitMigrations(conf, logger)
		}
		return postgresbd.NewPGDB(conf, logger)
	}
	if conf.FilePath != "" {
		logger.Infow("file config")
		return files.NewFile(conf, logger)
	}
	logger.Infow("memory config")
	return NewMapStorage()
}

func (s *MapStorage) Set(key, value string) bool {
	s.data[key] = value
	return true
}

func (s *MapStorage) Get(key string) (string, bool) {
	value, exist := s.data[key]
	return value, exist
}
