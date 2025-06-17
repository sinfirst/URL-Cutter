package models

// OriginalURL cтруктура необходимая для декодирования JSON
type OriginalURL struct {
	URL string `json:"url"`
}

// ResultURL cтруктура необходимая для кодирования JSON
type ResultURL struct {
	Result string `json:"result"`
}

// ShortenRequestForBatch cтруктура необходимая для декодирования JSON при запросе на группу урлов
type ShortenRequestForBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ShortenRequestForBatch cтруктура необходимая для кодирования JSON при запросе на группу урлов
type ShortenResponceForBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ShortenOrigURLs cтруктура необходимая для получения иноформации о всех урлах определнного пользователя
type ShortenOrigURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
