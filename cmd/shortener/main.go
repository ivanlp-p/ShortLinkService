package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ivanlp-p/ShortLinkService/cmd/config"
	"github.com/ivanlp-p/ShortLinkService/internal/compress"
	"github.com/ivanlp-p/ShortLinkService/internal/logger"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
	"github.com/ivanlp-p/ShortLinkService/internal/storage"
	"github.com/ivanlp-p/ShortLinkService/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"strings"
)

func handler(storage storage.Storage, conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Info("This is handler")
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		originalURL := strings.TrimSpace(string(body))
		shortID := utils.ShortenURL(originalURL)
		shortLink := models.ShortLink{
			UUID:        uuid.NewString(),
			ShortURL:    shortID,
			OriginalURL: originalURL,
		}

		err = storage.PutOriginalURL(context.Background(), shortLink)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				if existingShort, found, err := storage.GetShortURLByOriginalURL(context.Background(), originalURL); err == nil && found {
					shortURL := fmt.Sprintf(conf.BaseURL+"%s", existingShort)
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte(shortURL))
					return
				}
			}

			http.Error(w, "Failed to save", http.StatusInternalServerError)
			return
		}
		shortURL := fmt.Sprintf(conf.BaseURL+"%s", shortID)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func handlerGet(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Info("This is handlerGet")
		fmt.Println("ID: ", chi.URLParam(r, "id"))
		logger.Log.Info("ID: ", zap.String("id - ", chi.URLParam(r, "id")))
		id := chi.URLParam(r, "id")

		originalURL, err := storage.GetOriginalURL(context.Background(), id)

		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		_, err = io.ReadAll(r.Body)
		if err != nil || originalURL == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		fmt.Println("originalURL: ", originalURL)

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func PostShortenRequest(storage storage.Storage, conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var originURL models.OriginalURL

		if err := json.NewDecoder(r.Body).Decode(&originURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		url := originURL.URL
		shortID := utils.ShortenURL(url)
		shortLink := models.ShortLink{UUID: uuid.NewString(),
			ShortURL:    shortID,
			OriginalURL: url,
		}

		if err := storage.PutOriginalURL(context.Background(), shortLink); err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				if existingShort, found, err := storage.GetShortURLByOriginalURL(context.Background(), originURL.URL); err == nil && found {
					shortURL := fmt.Sprintf(conf.BaseURL+"%s", existingShort)
					resp := models.ShortURL{
						Result: shortURL,
					}

					response, err := json.MarshalIndent(resp, "", "   ")
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusConflict)
					w.Write(response)
					return
				}
			}

			http.Error(w, "Failed to save", http.StatusInternalServerError)
			return
		}
		shortURL := conf.BaseURL + shortID

		resp := models.ShortURL{
			Result: shortURL,
		}

		response, err := json.MarshalIndent(resp, "", "   ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(response)
	}
}

func HandlerPing(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := storage.Ping(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandlerShortenBatch(storage storage.Storage, conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var batch []models.BatchRequest
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if len(batch) == 0 {
			http.Error(w, "Empty batch", http.StatusBadRequest)
			return
		}

		records := make([]models.ShortLink, 0, len(batch))
		responses := make([]models.BatchResponse, 0, len(batch))

		for _, item := range batch {
			shortURL := utils.ShortenURL(item.OriginalURL)
			UUID := uuid.NewString()

			records = append(records, models.ShortLink{
				UUID:        UUID,
				ShortURL:    shortURL,
				OriginalURL: item.OriginalURL,
			})
			responses = append(responses, models.BatchResponse{
				CorrelationID: item.CorrelationID,
				ShortURL:      conf.BaseURL + shortURL,
			})
		}

		if err := storage.BatchInsert(context.Background(), records); err != nil {
			http.Error(w, "Failed to save batch", http.StatusInternalServerError)
			return
		}

		respJSON, err := json.Marshal(responses)
		if err != nil {
			http.Error(w, "Encoding error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(respJSON)
	}
}

func main() {
	conf := config.Init()
	if err := initLogger(conf.LogLevel); err != nil {
		log.Fatal(err)
	}

	logger.Log.Info("Running server on", zap.String("Address", conf.Address))

	strg := initStorage(conf)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", logger.RequestLogger(compress.GzipCompress(handlerGet(strg))))
		r.Post("/", logger.RequestLogger(compress.GzipCompress(handler(strg, conf))))
		r.Route("/api/", func(r chi.Router) {
			r.Post("/shorten", logger.RequestLogger(compress.GzipCompress(PostShortenRequest(strg, conf))))
			r.Route("/shorten/", func(r chi.Router) {
				r.Post("/batch", logger.RequestLogger(compress.GzipCompress(HandlerShortenBatch(strg, conf))))
			})
		})
		r.Get("/ping", logger.RequestLogger(compress.GzipCompress(HandlerPing(strg))))
	})

	chi.Walk(r, func(method, route string, h http.Handler, m ...func(http.Handler) http.Handler) error {
		fmt.Printf("%-6s %s\n", method, route)
		return nil
	})

	log.Fatal(http.ListenAndServe(conf.Address, r))
}

func initStorage(conf *config.Config) storage.Storage {
	var strg storage.Storage
	var err error

	store := storage.NewMapStorage()

	if conf.DB != "" {
		strg, err = storage.NewPostgresStorage(context.Background(), conf.DB)
		if err != nil {
			logger.Log.Error("Failed to initialize PostgreSQL storage: %v. Falling back to file storage", zap.Error(err))
		}
	}

	if strg == nil && conf.FileStorage != "" {
		strg = storage.NewFileStorage(conf.FileStorage, store)
		err := strg.LoadFromFile()
		if err != nil {
			logger.Log.Error("Store not load")
		}
	}

	if strg == nil {
		strg = store
	}

	return strg
}

func initLogger(logLevel string) error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	return nil
}
