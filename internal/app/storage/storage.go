package storage

import (
	"context"
	"fmt"

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

type ShortenOrigURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage interface {
	SetURL(ctx context.Context, key, value string, userID int) error
	GetURL(ctx context.Context, key string) (string, error)
	GetWithUserID(ctx context.Context, UserID int) (map[string]string, error)
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
		postgresbd.InitMigrations(conf, logger)
		return postgresbd.NewPGDB(conf, logger)
	}
	if conf.FilePath != "" {
		logger.Infow("file config")
		return files.NewFile(conf, logger)
	}
	logger.Infow("memory config")
	return NewMapStorage()
}

func (s *MapStorage) SetURL(ctx context.Context, key, value string, userId int) error {
	s.data[key] = value
	return nil
}

func (s *MapStorage) GetURL(ctx context.Context, key string) (string, error) {
	value, flag := s.data[key]
	if flag {
		return value, nil
	}
	return value, fmt.Errorf("not found in storage")
}
func (s *MapStorage) GetWithUserID(ctx context.Context, UserID int) (map[string]string, error) {
	return nil, nil
}
