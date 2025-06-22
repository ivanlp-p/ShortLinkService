package storage

import (
	"context"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
)

type Storage interface {
	LoadFromFile() error
	PutOriginalURL(ctx context.Context, shortLink models.ShortLink) error
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	BatchInsert(ctx context.Context, links []models.ShortLink) error
	GetShortURLByOriginalURL(ctx context.Context, originalURL string) (string, bool, error)
	Ping(ctx context.Context) error
	Close() error
}
