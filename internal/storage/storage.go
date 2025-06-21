package storage

import (
	"context"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
)

type Storage interface {
	LoadFromFile() error
	PutOriginalURL(ctx context.Context, shortLink models.ShortLink) error
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	Ping(ctx context.Context) error
	Close() error
}
