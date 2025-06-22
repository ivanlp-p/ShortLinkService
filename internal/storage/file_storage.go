package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
	"os"
	"sync"
)

type FileStorage struct {
	fileName string
	store    *MapStorage
	mx       sync.RWMutex
}

func NewFileStorage(fileName string, store *MapStorage) Storage {
	return &FileStorage{
		fileName: fileName,
		store:    store,
	}
}

func (fs *FileStorage) LoadFromFile() error {
	fs.mx.Lock()
	defer fs.mx.Unlock()

	file, err := os.Open(fs.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // файл пока не существует — это ок
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record models.ShortLink
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return err
		}

		err = fs.store.PutOriginalURL(context.Background(), record)
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (fs *FileStorage) PutOriginalURL(ctx context.Context, shortLink models.ShortLink) error {
	fs.mx.Lock()
	defer fs.mx.Unlock()

	file, err := os.OpenFile(fs.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonLine, err := json.Marshal(shortLink)
	if err != nil {
		return err
	}
	_, err = file.WriteString(string(jsonLine) + "\n")
	if err != nil {
		return err
	}

	err = fs.store.PutOriginalURL(ctx, shortLink)
	if err != nil {
		return err
	}
	return err
}

func (fs *FileStorage) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	originalURL, err := fs.store.GetOriginalURL(ctx, shortURL)
	if err != nil {
		return "", errors.New("original URL not found")
	}

	return originalURL, nil
}

func (fs *FileStorage) BatchInsert(ctx context.Context, links []models.ShortLink) error {
	for _, item := range links {
		err := fs.PutOriginalURL(ctx, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (fs *FileStorage) Close() error {
	return nil
}
