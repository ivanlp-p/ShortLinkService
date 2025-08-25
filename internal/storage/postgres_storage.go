package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, pool *pgxpool.Pool) (*PostgresStorage, error) {

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	if err := createTables(ctx, pool); err != nil {
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
		`CREATE TABLE IF NOT EXISTS urls (
    uuid TEXT primary key,
     user_id TEXT NOT NULL,
      short_url TEXT NOT NULL UNIQUE,
      original_url TEXT NOT NULL);
`)

	if err != nil {
		return fmt.Errorf("unable to acquire connection: %w", err)
	}

	_, err = conn.Exec(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url_unique ON urls (original_url);`)
	if err != nil {
		return fmt.Errorf("failed to create unique index: %w", err)
	}

	return err
}

func (p PostgresStorage) LoadFromFile() error {
	return nil
}

func (p *PostgresStorage) PutOriginalURL(ctx context.Context, shortLink models.ShortLink) error {
	query := `INSERT INTO urls (uuid, user_id, short_url, original_url) VALUES ($1, $2, $3, $4)`
	_, err := p.pool.Exec(ctx, query, shortLink.UUID, shortLink.UserID, shortLink.ShortURL, shortLink.OriginalURL)

	return err
}

func (p PostgresStorage) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	var originalURL string
	err := p.pool.QueryRow(ctx, `SELECT original_url FROM urls WHERE short_url = $1`, shortURL).Scan(&originalURL)

	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("URL not found")
	}

	return originalURL, err
}

func (p PostgresStorage) BatchInsert(ctx context.Context, links []models.ShortLink) error {
	if len(links) == 0 {
		return nil
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	stmt := `INSERT INTO urls (uuid, user_id, short_url, original_url)
	         VALUES ($1, $2, $3, $4)
	         ON CONFLICT (short_url) DO NOTHING`

	for _, item := range links {
		_, err = tx.Exec(ctx, stmt, item.UUID, item.UserID, item.ShortURL, item.OriginalURL)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	return err
}

func (p PostgresStorage) GetShortURLByOriginalURL(ctx context.Context, originalURL string) (string, bool, error) {
	var shortURL string
	err := p.pool.QueryRow(ctx, `SELECT short_url FROM urls WHERE original_url = $1`, originalURL).Scan(&shortURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, fmt.Errorf("URL not found")
		}
		return "", false, err
	}

	return shortURL, true, err
}

func (p PostgresStorage) GetUrlsByUserID(ctx context.Context, userID string) ([]models.ShortLink, error) {
	var links []models.ShortLink

	rows, err := p.pool.Query(ctx, `SELECT uuid, short_url, original_url FROM urls WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var link models.ShortLink
		link.UserID = userID
		if err := rows.Scan(&link.UUID, &link.ShortURL, &link.OriginalURL); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, rows.Err()
}

func (p PostgresStorage) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p PostgresStorage) Close() error {
	_ = p.pool.Close
	return nil
}
