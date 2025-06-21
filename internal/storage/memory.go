package storage

import (
	"context"
	"fmt"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
	"sync"
)

type MapStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		data: make(map[string]string),
	}
}

func (s *MapStorage) PutOriginalURL(ctx context.Context, shortLink models.ShortLink) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[shortLink.ShortURL] = shortLink.OriginalURL

	return nil
}

func (s *MapStorage) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	url, exists := s.data[shortURL]
	if !exists {
		return "", fmt.Errorf("original URL not found")
	}

	return url, nil
}

func (s *MapStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MapStorage) Close() error {
	return nil
}

func (s *MapStorage) LoadFromFile() error {
	return nil
}
