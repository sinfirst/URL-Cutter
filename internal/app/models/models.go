package models

import "context"

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
	GetByUserID(ctx context.Context, userID int) (map[string]string, error)
}
