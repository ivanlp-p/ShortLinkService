package storage

import (
	"context"
	"fmt"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, dsn string) (Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	if err = createTables(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PostgresStorage{pool: pool}, nil
}

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("unable to acquire connection: %w", err)
	}

	defer conn.Release()

	_, err = conn.Exec(ctx,
		`CREATE TABLE IF NOT EXISTS urls (uuid TEXT primary key, short_url TEXT NOT NULL UNIQUE,original_url TEXT NOT NULL);`)
	if err != nil {
		return fmt.Errorf("unable to acquire connection: %w", err)
	}

	return err
}

func (p PostgresStorage) LoadFromFile() error {
	return nil
}

func (p PostgresStorage) PutOriginalURL(ctx context.Context, shortLink models.ShortLink) error {
	_, err := p.pool.Exec(ctx,
		`insert into urls (uuid, short_url, original_url) values ($1,$2,$3) on conflict (short_url) Do nothing`,
		shortLink.UUID, shortLink.ShortURL, shortLink.OriginalURL)

	return err
}

func (p PostgresStorage) GetOriginalURL(ctx context.Context, shortUrl string) (string, error) {
	var originalURL string
	err := p.pool.QueryRow(ctx, `SELECT original_url FROM urls WHERE short_url = $1`, shortUrl).Scan(&originalURL)

	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("URL not found")
	}

	return originalURL, err
}

func (p PostgresStorage) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p PostgresStorage) Close() error {
	p.Close()
	return nil
}
