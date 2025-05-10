package memory

import (
	"context"
	"fmt"
)

type MapStorage struct {
	data map[string]string
}

func NewMapStorage() *MapStorage {
	return &MapStorage{data: make(map[string]string)}
}

func (s *MapStorage) SetURL(ctx context.Context, key, value string, userID int) error {
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
func (s *MapStorage) GetByUserID(ctx context.Context, userID int) (map[string]string, error) {
	return nil, nil
}
