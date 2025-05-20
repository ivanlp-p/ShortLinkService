package storage

import (
	"bufio"
	"encoding/json"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
	"os"
	"sync"
)

type FileStorage struct {
	fileName string
	store    *MapStorage
	mx       sync.RWMutex
}

func NewFileStorage(fileName string, store *MapStorage) *FileStorage {
	return &FileStorage{
		fileName: fileName,
		store:    store,
	}
}

func (fs *FileStorage) LoadFromFile() error {
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
		fs.store.Set(record.ShortURL, record.OriginalURL)
	}
	return scanner.Err()
}

func (fs *FileStorage) SaveShortLink(shortLink models.ShortLink) error {
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
	fs.store.Set(shortLink.ShortURL, shortLink.OriginalURL)
	return err
}
