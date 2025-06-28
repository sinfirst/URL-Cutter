package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/sinfirst/URL-Cutter/internal/app/models"
)

// MapStorage структура
type MapStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewMapStorage конструктор для структуры MapStorage
func NewMapStorage() *MapStorage {
	return &MapStorage{data: make(map[string]string), mu: sync.RWMutex{}}
}

// SetURL сохранение данных в память
func (s *MapStorage) SetURL(ctx context.Context, key, value string, userID int) error {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
	return nil
}

// GetURL извлечение данных из памяти
func (s *MapStorage) GetURL(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	value, flag := s.data[key]
	if flag {
		return value, nil
	}
	s.mu.RUnlock()
	return value, fmt.Errorf("not found in storage")
}

// GetByUserID заглушка для DataBase
func (s *MapStorage) GetByUserID(ctx context.Context, userID int) ([]models.ShortenOrigURLs, error) {
	return nil, nil
}
