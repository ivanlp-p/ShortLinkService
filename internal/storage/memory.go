package storage

import "sync"

type MapStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		data: make(map[string]string),
	}
}

func (s *MapStorage) Set(id string, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = url
}

func (s *MapStorage) Get(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	url, ok := s.data[id]
	return url, ok
}
